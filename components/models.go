package components

import (
	"XAI-liacs/BLADE_db/internal/database"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/sqlc-dev/pqtype"
)

type LLMRepresentative interface {
	GetHardware(object any) pqtype.NullRawMessage
	GetConfig(object any) json.RawMessage
}

type LLMRepresentativeImpl struct{}

func (LLMRepresentativeImpl) GetHardware(object any) pqtype.NullRawMessage {
	switch v := object.(type) {
	case database.Llm:
		return v.Hardware
	case database.GetLLMSolutionCountsRow:
		return v.Hardware
	default:
		panic("unsupported LLM type")
	}
}

func (LLMRepresentativeImpl) GetConfig(object any) json.RawMessage {
	switch v := object.(type) {
	case database.Llm:
		return v.Config
	case database.GetLLMSolutionCountsRow:
		return v.Config
	default:
		panic("unsupported LLM type")
	}
}

func renderMarkdown(message string) string {
	return message
	// var buf bytes.Buffer
	// if err := goldmark.Convert([]byte(message), &buf); err != nil {
	// 	return message
	// }
	// return buf.String()
}

func ParseLLM(llm any) (hardware, config map[string]any) {
	representative := LLMRepresentativeImpl{}

	hardwareData := representative.GetHardware(llm)

	configData := representative.GetConfig(llm)
	if hardwareData.Valid {
		if err := json.Unmarshal(hardwareData.RawMessage, &hardware); err != nil {
			hardware = nil
		}
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		config = nil
	}

	return hardware, config
}

func ParseMethod(llm database.Method) (config map[string]any) {
	config = make(map[string]any)
	if err := json.Unmarshal(
		llm.Config,
		&config,
	); err != nil {
		config = nil
	}
	return config
}

func CleanDictionary(dict map[string]any) map[string]string {
	returndict := make(map[string]string)
	if dict == nil {
		return returndict
	}
	for key, value := range dict {
		key := strings.ReplaceAll(key, "_", " ")
		key = strings.ReplaceAll(key, "cpu", "CPU")
		key = strings.ReplaceAll(key, "gpu", "GPU")
		key = strings.ReplaceAll(key, "os", "OS")
		if key == "memory" {
			returndict[key] = fmt.Sprintf("%v MB", value)
			continue
		}
		if value == nil {
			returndict[key] = "-"
		} else {
			returndict[key] = fmt.Sprintf("%v", value)
		}
	}
	return returndict
}

func getInitials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || unicode.IsSpace(r)
	})

	if len(words) == 0 {
		return ""
	}

	if len(words) >= 2 {
		return strings.ToUpper(
			string([]rune(words[0])[0]) +
				string([]rune(words[len(words)-1])[0]),
		)
	}

	runes := []rune(words[0])
	if len(runes) == 1 {
		return strings.ToUpper(string(runes[0]))
	}

	return strings.ToUpper(
		string(runes[0]) + string(runes[len(runes)-1]),
	)
}

// Generate the json for rendering plotly...
// returns a [2, *] array, where lineArray[0] are x axis, and lineArray[1] y axis.
//
//	 	   a [2, *] array, where scatter[0] are x axis, and lineArray[1] y axis for scatter plot
//	 	   a [*] array, where markers are markers, for plot, o for valid, x for invalid solutions.
//				error: An error objection for defining what is wrong with the data.
func getPlotArray(data []database.GetSolutionsForRunRow) (lineArray [][]any, scatter [][]float32, markers []string, names []string, err error) {
	lineArray = make([][]any, 0)
	scatter = make([][]float32, 0)
	markers = make([]string, 0)
	names = make([]string, 0)
	// Declare data for serialisation.
	plotData, err := getPlotData(data)
	if err != nil {
		return
	}

	x := make([]float32, 0)
	y := make([]float32, 0)

	for _, value := range plotData {
		x = append(x, float32(value.Generation))
		names = append(names, value.Name)
		if value.Fitness == nil {
			y = append(y, 0)
			markers = append(markers, "x")
		} else {
			y = append(y, *value.Fitness)
			markers = append(markers, "circle")
		}
	}

	scatter = append(scatter, x)
	scatter = append(scatter, y)

	// Empty x and y.
	lineX := make([]any, 0)
	lineY := make([]any, 0)

	for _, child := range plotData {
		for _, parentID := range child.Parents {
			parent, ok := plotData[parentID]
			if !ok {
				continue
			}

			lineX = append(lineX, child.Generation)
			if child.Fitness == nil {
				lineY = append(lineY, 0)
			} else {
				lineY = append(lineY, *child.Fitness)
			}

			lineX = append(lineX, parent.Generation)
			if parent.Fitness == nil {
				lineY = append(lineY, 0)
			} else {
				lineY = append(lineY, *parent.Fitness)
			}

			// Break line after each parent-child relationship.
			lineX = append(lineX, nil)
			lineY = append(lineY, nil)
		}
	}

	lineArray = append(lineArray, lineX)
	lineArray = append(lineArray, lineY)

	return
}

// Serialise data.
type plotlyData struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Generation int      `json:"generation"`
	Parents    []string `json:"parents"`
	Fitness    *float32 `json:"fitness"`
}

func getPlotData(solutions []database.GetSolutionsForRunRow) (plotData map[string]plotlyData, err error) {
	plotData = make(map[string]plotlyData)
	for _, solution := range solutions {
		parents := make([]string, 0)
		var fitness *float32 = nil
		name := "-"
		generation := 0

		for _, parent := range solution.Parents {
			parents = append(parents, parent.String())
		}

		if solution.Fitness.Valid {
			fitnessVal := make(map[string]float32)

			err := json.Unmarshal(solution.Fitness.RawMessage, &fitnessVal)
			if err != nil {
				return plotData, err
			}

			if len(fitnessVal) > 1 {
				return plotData, errors.New("Multi-Objective lineage plot not implemented yet..")
			}

			if value, ok := fitnessVal["default"]; ok {
				fitness = &value
			}
		}
		if solution.Name.Valid {
			name = solution.Name.String
		}
		if solution.Generation.Valid {
			generation = int(solution.Generation.Int32)
		}
		plotData[solution.ID.String()] = plotlyData{
			ID:         solution.ID.String(),
			Name:       name,
			Generation: generation,
			Parents:    parents,
			Fitness:    fitness,
		}
	}
	return plotData, nil
}
