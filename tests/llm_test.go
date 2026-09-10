package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func TestBaseTableLLM(t *testing.T) {
	state := config.GetContext("test")
	insertions := make([]database.CreateLLMParams, 0)
	// preTestClear(state)
	files_content := make([]map[string]any, 0)
	files, err := get_file_names("test_files/llm")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range files {
		content, err := read_file_aray_content(file)
		if err != nil {
			t.Fatal(err.Error())
		} else {
			files_content = append(files_content, content...)
		}
	}

	// Assert insertions don't allow copy of LLMs.
	for index, data := range files_content {
		config, err := json.Marshal(data["config"])
		if err != nil {
			t.Fatalf("Unable to serialise config data: %s", err.Error())
		}
		hardware, err := json.Marshal(data["hardware"])
		if err != nil {
			t.Fatalf("Unable to serialise hardware data: %s", err.Error())
		}
		row := database.CreateLLMParams{
			ID:       uuid.New(),
			Model:    data["model"].(string),
			Hardware: pqtype.NullRawMessage{RawMessage: hardware, Valid: len(hardware) != 0},
			Config:   config,
		}
		_, err = state.DB.CreateLLM(context.Background(), row)
		if err != nil {
			if !strings.Contains(files[index], "gemma_2") { // Guard against duplication.
				t.Fatalf("Unable to insert %v, into db: %s", row, err.Error())
			}
		} else {
			insertions = append(insertions, row)
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

}
