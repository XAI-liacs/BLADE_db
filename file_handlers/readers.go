package filehandlers

import (
	"anantashahane/BLADE_db/internal/database"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// Deserialise helper for all functions, takes file_path, data pointer to
// deserialise data to, and pushes data in provided pointer.
func deserialiseFile(file_path string, data any) error {
	dat, err := os.ReadFile(file_path)
	if err != nil {
		return errors.New("Unable to read file: " + file_path + " error: " + err.Error())
	}
	err = json.Unmarshal(dat, &data)
	if err != nil {
		return errors.New("Unable to unmarshal data, error: " + err.Error())
	}
	return nil
}

// Key checker for compatibility.
func containsKeys[K comparable, V any](m map[K]V, keys []K) bool {
	for _, key := range keys {
		if _, ok := m[key]; !ok {
			return false
		}
	}
	return true
}

/* Interface, to make sure that all function in file_handlers module have following function
 * implementable in `file_details.Read()`, which would populate the some_struct with associated
 * file's content
 */
type file_manager interface {
	Read(file_details FileDetails) (any, error)
}

type FileDetails struct {
	Root   string
	Suffix string
}

// Empty struct to implement LLM reader.
type LLM struct{}

/*
LLM reader, takes a LLM file `llm.json` from each run directory and generates CreateLLMParams struct for DB injestion.

Params:

- `file_details`: FileDetails struct, with suffix = "llm.json".

Returns:

- database.CreatLLMParams: sqlc generated struct for insertion into DB.
- err: error message if something went wrong.
*/
func (llm LLM) Read(file_details FileDetails) (content []database.CreateLLMParams, err error) {
	content = make([]database.CreateLLMParams, 0)
	path := filepath.Join(file_details.Root, file_details.Suffix)
	json_data := make([]map[string]any, 0)
	err = deserialiseFile(path, &json_data)
	if err != nil {
		return content, err
	}
	for _, dict := range json_data {
		if !containsKeys(dict, []string{"model", "config"}) {
			return content, errors.New("llm.json must contain model, config, and optional(hardware) keys.")
		}
	}
	for _, llm_data := range json_data {
		row := database.CreateLLMParams{
			ID:    uuid.New(),
			Model: llm_data["model"].(string),
		}

		config, err := json.Marshal(llm_data["config"])
		if err != nil {
			return content, errors.New("Unable to marshal config, error: " + err.Error())
		}
		row.Config = config

		if hardware, ok := llm_data["hardware"]; ok {
			hardware_data, err := json.Marshal(hardware)
			if err != nil {
				return content, errors.New("Unable to marhsal hardware, error : " + err.Error())
			}
			row.Hardware = pqtype.NullRawMessage{RawMessage: hardware_data, Valid: true}
		} else {
			row.Hardware = pqtype.NullRawMessage{RawMessage: make([]byte, 0), Valid: false}
		}
		content = append(content, row)
	}
	return
}

// Empty struct to implement Method reader.
type Method struct{}

/*
Method reader, takes a Method file `method.json` from each run directory and generates CreateMethodParams struct for DB injestion.

Params:

- `file_details`: FileDetails struct, with suffix = "llm.json".

Returns:

- database.CreateMethodParams: sqlc generated struct for insertion into DB.
- err: error message if something went wrong.
*/
func (m Method) Read(file_details FileDetails) (content database.CreateMethodParams, err error) {
	path := filepath.Join(file_details.Root, file_details.Suffix)
	json_data := make(map[string]any, 0)
	err = deserialiseFile(path, &json_data)
	if err != nil {
		return content, err
	}
	if !containsKeys(json_data, []string{"name", "source", "config"}) {
		return content, errors.New("method.json must contain name, source, config, keys.")
	}
	content.ID = uuid.New()
	content.Name = json_data["name"].(string)
	content.Source = json_data["source"].(string)

	config, err := json.Marshal(json_data["config"])
	if err != nil {
		return content, errors.New("Unable to marshal config, error: " + err.Error())
	}
	content.Config = config
	return
}

// Empty struct to implement Problem reader.
type Problem struct{}

// On DB level, problem is divided into 2 tables, tags, and problems.
// So we implement a struct that has both, which is returned here.

type ReadProblemContent struct {
	Tags    []database.CreateTagParams
	Problem database.CreateProblemParams
}

/*
Problem reader, takes a problem file `problem.json` from each run directory and generates ReadProblemContent struct for DB injestion.

Params:

- `file_details`: FileDetails struct, with suffix = "problem.json".

Returns:

- ReadProblemContent: embedded with sqlc generated structs for insertion into DB.
- err: error message if something went wrong.
*/
func (p Problem) Read(file_details FileDetails) (content ReadProblemContent, err error) {
	path := filepath.Join(file_details.Root, file_details.Suffix)
	json_data := make(map[string]any, 0)
	err = deserialiseFile(path, &json_data)
	if err != nil {
		return content, err
	}
	if !containsKeys(json_data, []string{"tags", "name", "prompt", "minimisation", "evaluator", "config"}) {
		return content, errors.New("problem.json must contain tags, name, prompt, minimisation, evaluator, config")
	}
	db_tags := make([]database.CreateTagParams, 0)
	for _, tag := range json_data["tags"].([]any) {
		db_tags = append(db_tags, database.CreateTagParams{ID: uuid.New(), Tag: tag.(string)})
	}

	problem := database.CreateProblemParams{
		ID:           uuid.New(),
		Name:         json_data["name"].(string),
		Prompt:       json_data["prompt"].(string),
		Minimisation: json_data["minimisation"].(bool),
		Evaluator:    json_data["evaluator"].(string),
	}

	config, err := json.Marshal(json_data["config"])
	if err != nil {
		return content, errors.New("Unable to marshal config, error: " + err.Error())
	}
	problem.Config = pqtype.NullRawMessage{RawMessage: config, Valid: true}
	content.Tags = db_tags
	content.Problem = problem
	return
}
