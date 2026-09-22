package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// Helper functions:

func preTestClear(dbConfig config.State) error {
	fmt.Println("\t=== Clearing all tables ===")
	ctx := context.Background()
	// Clear Relation Tables First.....
	if err := dbConfig.DB.ClearRunDescriptor(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearConversationLog(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearMethodLLM(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearTagProblem(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearExperimentRun(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearParentChild(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearRunSolution(ctx); err != nil {
		return err
	}

	// Clear Base Tables Later.....
	if err := dbConfig.DB.ClearTag(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearRun(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearLLM(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearProblem(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearMessage(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearExperiment(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearMethod(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearSolution(ctx); err != nil {
		return err
	}
	return nil
}

func optional_string_equality(lhs, rhs sql.NullString) bool {
	if lhs.Valid == rhs.Valid {
		if !lhs.Valid {
			return true
		}
		return lhs.String == rhs.String
	}
	return false
}

func optional_int_equality(lhs, rhs sql.NullInt32) bool {
	if lhs.Valid == rhs.Valid {
		if !lhs.Valid {
			return true
		}
		return lhs.Int32 == rhs.Int32
	}
	return false
}

func getRandomTime(before time.Time) time.Time {
	date := before.AddDate(
		-rand.Intn(4),
		-rand.Intn(12),
		-rand.Intn(30))
	return date.In(time.Local)
}

func Equal(
	insertion database.CreateExperimentParams,
	retrieved database.Experiment,
) bool {
	if insertion.ID != retrieved.ID {
		return false
	}

	loc := insertion.StartDate.Location()

	if !insertion.StartDate.Equal(retrieved.StartDate.In(loc)) {
		return false
	}

	if !insertion.EndDate.Equal(retrieved.EndDate.In(loc)) {
		return false
	}

	return true
}

func getRandomExperiment() database.CreateExperimentParams {
	start_date := getRandomTime(time.Now())
	end_date := getRandomTime(start_date)
	experiment_name := []string{"Gravitation Wave Detection Aparatus Optimisation", "Steiner Tree Problem", "Auto-Correlation Inequality 3", "AutoML"}

	return database.CreateExperimentParams{
		ID:        uuid.New(),
		Name:      experiment_name[rand.Intn(len(experiment_name))],
		StartDate: start_date,
		EndDate:   end_date}
}

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

func TestMain(m *testing.M) {
	// setup — runs before the tests
	state := config.GetContext("test")
	if err := preTestClear(state); err != nil {
		panic("Error clearing tables: " + err.Error())
	}

	code := m.Run()
	os.Exit(code)
}
