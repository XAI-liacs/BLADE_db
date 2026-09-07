package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestMethodGetterAndSetter(t *testing.T) {
	state := config.GetContext("test")
	insertions := make([]database.CreateMethodParams, 0)
	// preTestClear(state)
	files_content := make([]map[string]any, 0)
	files, err := get_file_names("test_files/method")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range files {
		content, err := read_file(file)
		if err != nil {
			t.Fatal(err.Error())
		} else {
			files_content = append(files_content, content)
		}
	}

	// Assert insertions don't allow copy of Methods.
	for index, data := range files_content {
		config, err := json.Marshal(data["config"])
		if err != nil {
			t.Fatalf("Unable to serialise config data: %s", err.Error())
		}

		row := database.CreateMethodParams{
			ID:     uuid.New(),
			Name:   data["name"].(string),
			Source: data["source"].(string),
			Config: config,
		}
		_, err = state.DB.CreateMethod(context.Background(), row)
		if err != nil {
			if !strings.Contains(files[index], "llamea2") { // Guard against duplication.
				t.Fatalf("Unable to insert %v, into db: %s", row, err.Error())
			}
		} else {
			insertions = append(insertions, row)
		}

		// Fetch check if data is inserted properly:
		feteched_methods, err := state.DB.GetMethods(context.Background())
		if err != nil {
			t.Fatal("Unable to perform fetch on database..")
		}
		for index, fetched_llm := range feteched_methods {
			if insertions[index].ID != fetched_llm.ID {
				t.Fatalf("Mismatched ID: %s vs %s", insertions[index].ID.String(), fetched_llm.ID.String())
			}
			if insertions[index].Name != fetched_llm.Name {
				t.Fatalf("Mismatched Model: %s vs %s", insertions[index].Name, fetched_llm.Name)
			}
			if insertions[index].Source != fetched_llm.Source {
				t.Fatalf("Mismatched Hardware: \n%s \nvs\n %s", insertions[index].Source, fetched_llm.Source)
			}
			if !isEqualJsonObjects(insertions[index].Config, fetched_llm.Config) {
				t.Fatalf("Mismatched Config: \n%s \nvs\n %s", insertions[index].Config.String(), fetched_llm.Config.String())
			}
		}
	}

}
