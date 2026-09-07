package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func get_file_names(base string) ([]string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{}, err
	}
	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			fileNames = append(fileNames, fmt.Sprintf("%s/%s",
				base, entry.Name()),
			)

		}
	}
	return fileNames, nil
}

func read_file(file_name string) (content map[string]any, err error) {
	content_byte, err := os.ReadFile(file_name)
	if err != nil {
		return content, errors.New("Cannot read file " + file_name + " error: " + err.Error())
	}
	err = json.Unmarshal(content_byte, &content)
	if err != nil {
		return content, errors.New("Cannot deserialise json file " + file_name + " error: " + err.Error())
	}
	return content, err
}

func read_file_aray_content(file_name string) (content []map[string]any, err error) {
	content_byte, err := os.ReadFile(file_name)
	if err != nil {
		return content, errors.New("Cannot read file " + file_name + " error: " + err.Error())
	}
	err = json.Unmarshal(content_byte, &content)
	if err != nil {
		return content, errors.New("Cannot deserialise json file " + file_name + " error: " + err.Error())
	}
	return content, err
}

func isEqualNullableJsonObject(m1 pqtype.NullRawMessage, m2 pqtype.NullRawMessage) bool {
	if m1.Valid != m2.Valid {
		return false
	}
	if !m1.Valid && !m2.Valid {
		return true
	}
	data1 := make(map[string]any)
	data2 := make(map[string]any)
	err := json.Unmarshal(m1.RawMessage, &data1)
	if err != nil {
		return false
	}
	err = json.Unmarshal(m2.RawMessage, &data2)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(data1, data2)
}

func isEqualJsonObjects(m1 json.RawMessage, m2 json.RawMessage) bool {
	data1 := make(map[string]any)
	data2 := make(map[string]any)
	err := json.Unmarshal(m1, &data1)
	if err != nil {
		return false
	}
	err = json.Unmarshal(m2, &data2)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(data1, data2)
}

func TestFilesAvailable(t *testing.T) {
	file_names, err := get_file_names("test_files/problems")
	if err != nil {
		t.Fatalf("Cannot find problem files %s", err.Error())
	}
	fmt.Println("Found the following files: ")
	for _, file_name := range file_names {
		fmt.Println(file_name)
	}
}

func TestMessageSetterAndGetter(t *testing.T) {
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
		fmt.Printf("------------------------%d: %s------------------------\n", index, file_name)
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
