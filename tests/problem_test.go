package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func TestHelperFunctionFilesAvailable(t *testing.T) {
	file_names, err := get_file_names("test_files/problems")
	if err != nil {
		t.Fatalf("Cannot find problem files %s", err.Error())
	}
	fmt.Println("Found the following files: ")
	for _, file_name := range file_names {
		fmt.Println(file_name)
	}
}

func TestBaseTableProblem(t *testing.T) {
	state := config.GetContext("test")
	// preTestClear(state)
	file_names, err := get_file_names("test_files/problems")
	if err != nil {
		t.Fatalf("Cannot find problem files %s", err.Error())
	}
	// insertion_ids := make([]string, 0, 3)
	insertions := make([]database.CreateProblemParams, 0, 3)
	configs := make([]map[string]any, 0)
	for index, file_name := range file_names {
		content, err := read_file(file_name)
		if err != nil {
			t.Fatal(err.Error())
		}

		data := make([]byte, 0)
		if len(content["config"].(map[string]any)) != 0 {
			data, err = json.Marshal(content["config"])
			if err != nil {
				t.Fatalf("Cannot marshal config into data, error: %s", err.Error())
			}
		}

		insertion := database.CreateProblemParams{
			ID:           uuid.New(),
			Name:         content["name"].(string),
			Prompt:       content["prompt"].(string),
			Minimisation: content["minimisation"].(bool),
			Evaluator:    content["evaluator"].(string),
			Config:       pqtype.NullRawMessage{RawMessage: data, Valid: len(data) > 0},
		}
		_, err = state.DB.CreateProblem(context.Background(), insertion)
		if err != nil {
			if !strings.Contains(file_names[index], "Auto_corr_2_problem_2") { //Duplicate entry must fail.
				t.Fatalf("Got error inserting %s: %s", file_names[index], err.Error())
			}
		} else {
			insertions = append(insertions, insertion)
			configs = append(configs, content["config"].(map[string]any))
		}
	}

	// Test getting.
	retrieved_messages, err := state.DB.GetProblems(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}
	for index, inserted_item := range retrieved_messages {
		if insertions[index].ID != inserted_item.ID {
			t.Fatalf("ID miss match %v vs %v", insertions[index].ID, inserted_item.ID)
		}
		if insertions[index].Name != inserted_item.Name {
			t.Fatalf("Name miss match %v vs %v", insertions[index].Name, inserted_item.Name)
		}
		if insertions[index].Prompt != inserted_item.Prompt {
			t.Fatalf("Prompt miss match %v vs %v", insertions[index].Prompt, inserted_item.Prompt)
		}
		if insertions[index].Minimisation != inserted_item.Minimisation {
			t.Fatalf("Minimisation miss match %v vs %v", insertions[index].Minimisation, inserted_item.Minimisation)
		}
		if insertions[index].Evaluator != inserted_item.Evaluator {
			t.Fatalf("Evaluator miss match %v vs %v", insertions[index].Evaluator, inserted_item.Evaluator)
		}
		if !isEqualNullableJsonObject(insertions[index].Config, inserted_item.Config) {
			t.Fatalf("Config miss match %v vs %v", insertions[index].Config, inserted_item.Config)
		}
	}
}
