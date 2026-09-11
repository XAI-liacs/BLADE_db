package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type TestSolution struct {
	input          database.CreateSolutionParams
	expect_success bool
}

func generate_test_sequence() (test_instance []TestSolution, err error) {
	fitness := make(map[string]float32)
	fitness["fitness"] = 5

	test_solutions := make([]TestSolution, 0, 3)
	data, err := json.Marshal(fitness)
	if err != nil {
		return make([]TestSolution, 0), err
	}
	input1 := database.CreateSolutionParams{
		ID:          uuid.New(),
		Name:        sql.NullString{String: "SwiftGreeter", Valid: true},
		Description: sql.NullString{String: "A simple string interpolating hello swift.", Valid: true},
		Generation:  sql.NullInt32{Int32: 0, Valid: true},
		Code: sql.NullString{String: `func Greet(name: String) -> String {
	return "Hello, \(name)."
}`, Valid: true},
		Metadata: pqtype.NullRawMessage{RawMessage: make([]byte, 0), Valid: false},
		Fitness:  pqtype.NullRawMessage{RawMessage: data, Valid: false},
	}
	test_solutions = append(test_solutions, TestSolution{input: input1, expect_success: true})
	input2 := database.CreateSolutionParams{
		ID:          uuid.New(),
		Name:        sql.NullString{String: "PythonGreeter", Valid: true},
		Description: sql.NullString{String: "A simple string interpolating hello python.", Valid: true},
		Generation:  sql.NullInt32{Int32: 0, Valid: true},
		Code: sql.NullString{String: `def Greet(name: String) -> String:
	return f"Hello, {name}."`, Valid: true},
		Metadata: pqtype.NullRawMessage{RawMessage: make([]byte, 0), Valid: false},
		Fitness:  pqtype.NullRawMessage{RawMessage: data, Valid: false},
	}
	test_solutions = append(test_solutions, TestSolution{input: input2, expect_success: true})

	input3 := database.CreateSolutionParams{
		ID:          uuid.New(),
		Name:        sql.NullString{String: "GoGreeter", Valid: true},
		Description: sql.NullString{String: "A simple string interpolating hello go.", Valid: true},
		Generation:  sql.NullInt32{Int32: 0, Valid: true},
		Code: sql.NullString{String: `package main

import "fmt"

func Greet(name String) (String) {
	return fmt.Sprintf("Hello, %s.", name)
}`, Valid: true},
		Metadata: pqtype.NullRawMessage{RawMessage: make([]byte, 0), Valid: false},
		Fitness:  pqtype.NullRawMessage{RawMessage: data, Valid: false},
	}
	test_solutions = append(test_solutions, TestSolution{input: input3, expect_success: true})
	return test_solutions, nil
}

func TestBaseTableSolution(t *testing.T) {
	test_data, err := generate_test_sequence()
	if err != nil {
		t.Fatalf("Unable to instantiate test cases: %s", err.Error())
	}
	state := config.GetContext("test")

	for _, test_data := range test_data {
		_, err := state.DB.CreateSolution(context.Background(), test_data.input)
		if err != nil {
			if test_data.expect_success {
				t.Fatalf("Unable to insert into database: %s", err.Error())
			}
		} else {
		}
	}
	fetched_data, err := state.DB.GetSolutions(context.Background())
	if err != nil {
		t.Fatalf("Unable to fetch from dtabase: %s", err.Error())
	}
	for index, row := range fetched_data {
		if test_data[index].input.ID != row.ID {
			t.Fatalf("ID mis match %v vs %v", test_data[index].input.ID, row.ID)
		}
		if !optional_string_equality(test_data[index].input.Name, row.Name) {
			t.Fatalf("Name mis match %v vs %v", test_data[index].input.Name, row.Name)
		}
		if !optional_string_equality(test_data[index].input.Description, row.Description) {
			t.Fatalf("Description mis match %v vs %v", test_data[index].input.Description, row.Description)
		}
		if !optional_int_equality(test_data[index].input.Generation, row.Generation) {
			t.Fatalf("Generation mis match %v vs %v", test_data[index].input.Generation, row.Generation)
		}
		if !optional_string_equality(test_data[index].input.Code, row.Code) {
			t.Fatalf("Code mis match %v vs %v", test_data[index].input.Code, row.Code)
		}
		if !isEqualNullableJsonObject(test_data[index].input.Metadata, row.Metadata) {
			t.Fatalf("Metadata mis match %v vs %v", test_data[index].input.Metadata, row.Metadata)
		}
		if !isEqualNullableJsonObject(test_data[index].input.Fitness, row.Fitness) {
			t.Fatalf("Metadata mis match %v vs %v", test_data[index].input.Fitness, row.Fitness)
		}
	}
}
