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
)

func prepareMethod(state config.State) (insertions []database.CreateMethodParams, err error) {
	insertions = make([]database.CreateMethodParams, 0)
	// preTestClear(state)
	files_content := make([]map[string]any, 0)
	files, err := get_file_names("test_files/method")
	if err != nil {
		return
	}

	for _, file := range files {
		content, err := read_file(file)
		if err != nil {
			return insertions, err
		} else {
			files_content = append(files_content, content)
		}
	}

	// Assert insertions don't allow copy of Methods.
	for index, data := range files_content {
		config, err := json.Marshal(data["config"])
		if err != nil {
			return insertions, errors.New("Unable to serialise config data: " + err.Error())
		}

		row := database.CreateMethodParams{
			ID:     uuid.New(),
			Name:   data["name"].(string),
			Source: data["source"].(string),
			Config: config,
		}
		id, err := state.DB.CreateMethod(context.Background(), row)
		if err != nil {
			err_message := fmt.Sprintf("Unable to insert %v into db: %s", row, err.Error())
			return insertions, errors.New(err_message)
		} else if strings.Contains(files[index], "llamea2") { // Guard against duplication.
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

func TestBaseTableMethod(t *testing.T) {
	state := config.GetContext("test")
	// Assert insertions don't allow copy of Methods.
	insertions, err := prepareMethod(state)
	if err != nil {
		t.Fatal(err.Error())
	}
	// Fetch check if data is inserted properly:
	feteched_methods, err := state.DB.GetMethods(context.Background())
	if err != nil {
		t.Fatal("Unable to perform fetch on database..")
	}
	for index, fetched_method := range feteched_methods {
		if insertions[index].ID != fetched_method.ID {
			t.Fatalf("Mismatched ID: %s vs %s", insertions[index].ID.String(), fetched_method.ID.String())
		}
		if insertions[index].Name != fetched_method.Name {
			t.Fatalf("Mismatched Model: %s vs %s", insertions[index].Name, fetched_method.Name)
		}
		if insertions[index].Source != fetched_method.Source {
			t.Fatalf("Mismatched Hardware: \n%s \nvs\n %s", insertions[index].Source, fetched_method.Source)
		}
		if !isEqualJsonObjects(insertions[index].Config, fetched_method.Config) {
			t.Fatalf("Mismatched Config: \n%s \nvs\n %s", insertions[index].Config.String(), fetched_method.Config.String())
		}
	}
}
