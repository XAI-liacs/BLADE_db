package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func prepareLLM(state config.State) (insertions []database.CreateLLMParams, err error) {
	insertions = make([]database.CreateLLMParams, 0)
	// preTestClear(state)
	files_content := make([]map[string]any, 0)
	files, err := get_file_names("test_files/llm")
	if err != nil {
		return insertions, err
	}

	for _, file := range files {
		content, err := read_file_aray_content(file)
		if err != nil {
			return insertions, err
		} else {
			files_content = append(files_content, content...)
		}
	}

	// Assert insertions don't allow copy of LLMs.
	for index, data := range files_content {
		config, err := json.Marshal(data["config"])
		if err != nil {
			return insertions, errors.New("Unable to serialise config data: " + err.Error())
		}
		hardware, err := json.Marshal(data["hardware"])
		if err != nil {
			return insertions, errors.New("Unable to serialise hardware data: " + err.Error())
		}
		row := database.CreateLLMParams{
			ID:       uuid.New(),
			Model:    data["model"].(string),
			Hardware: pqtype.NullRawMessage{RawMessage: hardware, Valid: len(hardware) != 0},
			Config:   config,
		}
		id, err := state.DB.CreateLLM(context.Background(), row)
		if err != nil {
			err_message := fmt.Sprintf("Unable to insert %v into db: %s", row, err.Error())
			return insertions, errors.New(err_message)
		} else if strings.Contains(files[index], "gemma_2") { // Guard against duplication.
			if row.ID == id {
				err_message := fmt.Sprintf("Update on file %s changed ID (diff: %s -> %s). Old id's not expected to diff.", files[index], row.ID, id)
				return insertions, errors.New(err_message)
			}
		} else {
			insertions = append(insertions, row)
		}
	}
	return insertions, nil
}

func TestBaseTableLLM(t *testing.T) {
	state := config.GetContext("test")
	// Insertions test.
	insertions, err := prepareLLM(state)
	if err != nil {
		t.Fatal(err)
	}
	// Fetch check if data is inserted properly:
	fetched_llms, err := state.DB.GetLLMs(context.Background())
	if err != nil {
		t.Fatal("Unable to perform fetch on database..")
	}
	for index, fetched_llm := range fetched_llms {
		if insertions[index].ID != fetched_llm.ID {
			t.Fatalf("Mismatched ID: %s vs %s", insertions[index].ID.String(), fetched_llm.ID.String())
		}
		if insertions[index].Model != fetched_llm.Model {
			t.Fatalf("Mismatched Model: %s vs %s", insertions[index].Model, fetched_llm.Model)
		}
		if !isEqualNullableJsonObject(insertions[index].Hardware, fetched_llm.Hardware) {
			t.Fatalf("Mismatched Hardware: \n%s \nvs\n %s", insertions[index].Hardware.RawMessage.String(), fetched_llm.Hardware.RawMessage.String())
		}
		if !isEqualJsonObjects(insertions[index].Config, fetched_llm.Config) {
			t.Fatalf("Mismatched Config: \n%s \nvs\n %s", insertions[index].Config.String(), fetched_llm.Config.String())
		}
	}
}
