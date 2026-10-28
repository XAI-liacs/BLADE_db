package components

import (
	"XAI-liacs/BLADE_db/internal/database"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

func renderMarkdown(message string) string {
	return message
	// var buf bytes.Buffer
	// if err := goldmark.Convert([]byte(message), &buf); err != nil {
	// 	return message
	// }
	// return buf.String()
}

func ParseLLM(llm database.Llm) (hardware, config map[string]any) {
	if llm.Hardware.Valid {
		hardware = make(map[string]any)
		if err := json.Unmarshal(
			llm.Hardware.RawMessage,
			&hardware,
		); err != nil {
			hardware = nil
		}
	}
	config = make(map[string]any)
	if err := json.Unmarshal(
		llm.Config,
		&config,
	); err != nil {
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
	for key, value := range dict {
		key := strings.ReplaceAll(key, "_", " ")
		key = strings.ReplaceAll(key, "cpu", "CPU")
		key = strings.ReplaceAll(key, "gpu", "GPU")
		key = strings.ReplaceAll(key, "os", "OS")
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
				string([]rune(words[1])[0]),
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
