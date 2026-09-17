package filehandlers

import (
	"anantashahane/BLADE_db/internal/database"
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func OptionalUnwrap(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// Time translation to RFC3339
type CustomTime struct {
	time.Time
}

func (t *CustomTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	layouts := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02 15:04:05.999999",
	}

	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, s, time.UTC)
		if err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse time %q", s)
}

// Deserialise helper for all functions, takes file_path, data pointer to
// deserialise data to, and pushes data in provided pointer.
func deserialiseFile(filePath string, data any) error {
	if strings.HasSuffix(filePath, ".jsonl") {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", filePath, err)
		}
		defer file.Close()

		v := reflect.ValueOf(data)
		if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
			return errors.New("data must be a pointer to a slice for JSONL files")
		}

		slice := v.Elem()
		elemType := slice.Type().Elem()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			elem := reflect.New(elemType)

			if err := json.Unmarshal(scanner.Bytes(), elem.Interface()); err != nil {
				return fmt.Errorf("unable to unmarshal JSONL line: %w", err)
			}

			slice.Set(reflect.Append(slice, elem.Elem()))
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("unable to read file %s: %w", filePath, err)
		}

		return nil
	}

	dat, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %w", filePath, err)
	}

	if err := json.Unmarshal(dat, data); err != nil {
		return fmt.Errorf("unable to unmarshal data: %w", err)
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

// Struct to implement LLM,json parsing.
type LLM struct {
	Model    string         `json:"model"`
	Config   map[string]any `json:"config"`
	Hardware map[string]any `json:"hardware,omitempty"`
}

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
	llms := make([]LLM, 0)
	err = deserialiseFile(path, &llms)
	if err != nil {
		return content, err
	}
	// for _, llm := range llms {
	// 	if !containsKeys(dict, []string{"model", "config"}) {
	// 		return content, errors.New("llm.json must contain model, config, and optional(hardware) keys.")
	// 	}
	// }
	for _, llm_data := range llms {
		row := database.CreateLLMParams{
			ID:    uuid.New(),
			Model: llm_data.Model,
		}

		config, err := json.Marshal(llm_data.Config)
		if err != nil {
			return content, errors.New("Unable to marshal config, error: " + err.Error())
		}
		row.Config = config
		if len(llm_data.Hardware) != 0 {
			hardware_data, err := json.Marshal(llm_data.Hardware)
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

// Struct to implement Method.json reader.
type Method struct {
	Name   string         `json:"name"`
	Source string         `json:"source"`
	Config map[string]any `json:"config"`
}

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

	err = deserialiseFile(path, &m)
	if err != nil {
		return content, err
	}

	content.ID = uuid.New()
	content.Name = m.Name
	content.Source = m.Source

	config, err := json.Marshal(m.Config)
	if err != nil {
		return content, errors.New("Unable to marshal config, error: " + err.Error())
	}
	content.Config = config
	return
}

// Struct to implement Problem.json reader.
type Problem struct {
	ID           uuid.UUID      `json:"id"`
	Tags         []string       `json:"tags"`
	Name         string         `json:"name"`
	Prompt       string         `json:"prompt"`
	Minimisation bool           `json:"minimisation"`
	Evaluator    string         `json:"evaluator"`
	Config       map[string]any `json:"config"`
}

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
	err = deserialiseFile(path, &p)
	if err != nil {
		return content, err
	}
	db_tags := make([]database.CreateTagParams, 0)
	for _, tag := range p.Tags {
		db_tags = append(db_tags, database.CreateTagParams{ID: uuid.New(), Tag: tag})
	}

	problem := database.CreateProblemParams{
		ID:           uuid.New(),
		Name:         p.Name,
		Prompt:       p.Prompt,
		Minimisation: p.Minimisation,
		Evaluator:    p.Evaluator,
	}
	if p.ID == uuid.Nil {
		return content, errors.New("Unable to parse problem file correctly.")
	}

	config, err := json.Marshal(p.Config)
	if err != nil {
		return content, errors.New("Unable to marshal config, error: " + err.Error())
	}
	problem.Config = pqtype.NullRawMessage{RawMessage: config, Valid: true}
	content.Tags = db_tags
	content.Problem = problem
	return
}

// Struct to parse ConversationLog.jsonl files.
type ConversationLog struct {
	Role    string     `json:"role"`
	Time    CustomTime `json:"time"`
	Content string     `json:"content"`
	Cost    float32    `json:"cost"`
	Tokens  int        `json:"tokens"`
}

/*
ConversationLog reader, takes a conversationlog file `conversationlog.jsonl` from each run directory and generates []ConversationLog struct for DB injestion.

Params:

- `file_details`: FileDetails struct, with suffix = "conversationlog.jsonl".

Returns:

- []ConversationLog: A struct containing details for filling `message` and `Conversation_log` tables.
- err: error message if something went wrong.
*/
func (cl ConversationLog) Read(file_details FileDetails) (content []ConversationLog, err error) {
	path := filepath.Join(file_details.Root, file_details.Suffix)
	content = make([]ConversationLog, 0)
	err = deserialiseFile(path, &content)
	if err != nil {
		return content, err
	}
	for _, cl := range content {
		if cl.Role == "" || cl.Content == "" || cl.Time.IsZero() {
			return content, errors.New("Incorrectly parsed data.")
		}
	}
	return
}

// Struct to parse log.jsonl files.
type SolutionLog struct {
	ID          uuid.UUID      `json:"id"`
	Fitness     any            `json:"fitness"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Code        string         `json:"code"`
	Generation  *int           `json:"generation,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

/*
SolutionLog reader, takes a solution log file `log.jsonl` from each run directory and generates []CreateSolutionParams struct for DB injestion.

Params:

- `file_details`: FileDetails struct, with suffix = "log.jsonl".

Returns:

- []CreateSolutionParams: A sqlc generated struct for db injection
- err: error message if something went wrong.
*/
func (sl SolutionLog) Read(file_details FileDetails) (content []database.CreateSolutionParams, err error) {
	path := filepath.Join(file_details.Root, file_details.Suffix)
	content = make([]database.CreateSolutionParams, 0)
	data := make([]SolutionLog, 0)
	err = deserialiseFile(path, &data)
	if err != nil {
		return content, err
	}
	for _, sl := range data {
		metadata_json, err := json.Marshal(sl.Metadata)
		if err != nil {
			return content, errors.New("Unable to marshal Metadata: " + err.Error())
		}
		fitness_map := make(map[string]any)
		switch fitness := sl.Fitness.(type) {
		case float64:
			fitness_map["default"] = fitness
		case map[string]any:
			fitness_map = fitness
		case string:
			fitness_map = map[string]any{}
		}

		fitness_json, err := json.Marshal(fitness_map)
		if err != nil {
			return content, errors.New("Unable to marshal Fitness: " + err.Error())
		}

		row := database.CreateSolutionParams{
			ID:          sl.ID,
			Name:        sql.NullString{String: sl.Name, Valid: sl.Name != ""},
			Description: sql.NullString{String: sl.Description, Valid: sl.Description != ""},
			Generation:  sql.NullInt32{Int32: int32(OptionalUnwrap(sl.Generation)), Valid: sl.Generation != nil},
			Code:        sql.NullString{String: sl.Code, Valid: sl.Code != ""},
			Metadata:    pqtype.NullRawMessage{RawMessage: metadata_json, Valid: len(metadata_json) > 0},
			Fitness:     pqtype.NullRawMessage{RawMessage: fitness_json, Valid: len(fitness_json) > 0},
		}
		content = append(content, row)
	}
	return
}

// Struct to parse prgress.jsonl files.
type Run struct {
	MethodName   string     `json:"method_name"`
	ProblemName  string     `json:"problem_name"`
	Seed         int32      `json:"seed"`
	Budget       int        `json:"budget"`
	Evaluations  int        `json:"evaluation"`
	StartDate    CustomTime `json:"start_time"`
	EndDate      CustomTime `json:"end_time"`
	LogDirectory string     `json:"log_dir"`
}
type Progress struct {
	ID        uuid.UUID  `json:"id"`
	StartDate CustomTime `json:"start_time"`
	EndDate   CustomTime `json:"end_time"`
	Current   int        `json:"current"`
	Total     int        `json:"total"`
	Runs      []Run      `json:"runs"`
}

// Run data:
type ProgressData struct {
	runData      map[string]database.CreateRunParams //Map of directories, and its createRunparam
	progressData database.CreateExperimentParams
}

/*
SolutionLog reader, takes a solution log file `log.jsonl` from each run directory and generates []CreateSolutionParams struct for DB injestion.

Params:

- `file_details`: FileDetails struct, with suffix = "log.jsonl".

Returns:

- []CreateSolutionParams: A sqlc generated struct for db injection
- err: error message if something went wrong.
*/
func (p Progress) Read(file_details FileDetails) (content ProgressData, err error) {
	path := filepath.Join(file_details.Root, file_details.Suffix)
	err = deserialiseFile(path, &p)
	if err != nil {
		return content, err
	}

	if p.EndDate.IsZero() {
		return content, fmt.Errorf("Incomplete experiments are not allowed to be submitted: got End_date: %s", p.EndDate.Time)
	}

	name := filepath.Base(file_details.Root)
	content.progressData = database.CreateExperimentParams{
		ID:        p.ID,
		Name:      name,
		StartDate: p.StartDate.Time,
		EndDate:   p.EndDate.Time,
	}
	content.runData = make(map[string]database.CreateRunParams)
	for _, run := range p.Runs {
		run_data := database.CreateRunParams{
			ID:   uuid.New(),
			Name: fmt.Sprintf("%s-%s-%d", run.MethodName, run.ProblemName, run.Seed),
			Seed: run.Seed,
		}
		content.runData[run.LogDirectory] = run_data
	}
	return
}
