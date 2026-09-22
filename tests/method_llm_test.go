package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type MethodLLM struct {
	method_id uuid.UUID
	llm_id    uuid.UUID
}

func prepareMethodLLM(state config.State) (method_llm []MethodLLM, err error) {
	preTestClear(state)
	method_llm = make([]MethodLLM, 0)
	methods, err := prepareMethod(state)
	if err != nil {
		return method_llm, errors.New("Unable to prepare method table: " + err.Error())
	}
	llms, err := prepareLLM(state)
	if err != nil {
		return method_llm, errors.New("Unable to prepare method table: " + err.Error())
	}
	for method_index, method := range methods {
		for llm_index := method_index; llm_index < len(llms); llm_index++ {
			method_llm = append(method_llm, MethodLLM{
				method_id: method.ID,
				llm_id:    llms[llm_index].ID,
			})
		}
	}
	return method_llm, nil

}

func TestRelationTableMethodLLMInsertion(t *testing.T) {
	state := config.GetContext("test")
	testCases, err := prepareMethodLLM(state)
	if err != nil {
		t.Fatalf("Cannot prepare method and LLM tables: %s", err.Error())
	}

	// Insert all method llm relations.
	for _, testCase := range testCases {
		err = state.DB.CreateMethodLLM(
			context.Background(),
			database.CreateMethodLLMParams{
				MethodID: testCase.method_id,
				LlmID:    testCase.llm_id,
			})
		if err != nil {
			t.Fatalf(
				"Unable to insert into method_llm: %s",
				err,
			)
		}
	}

	// Check whether experiment run are valid.
	for _, testCase := range testCases {
		llms, err := state.DB.GetLLMsForMethod(
			context.Background(),
			testCase.method_id,
		)
		if err != nil {
			t.Fatalf(
				"Unable to get llms for method %s: %s",
				testCase.method_id,
				err,
			)
		}

		found := false
		for _, llm := range llms {
			if llm.ID == testCase.llm_id {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"Expected method %s to have llm %s",
				testCase.method_id.String(),
				testCase.llm_id.String(),
			)
		}
	}
}

func TestRelationTableMethodLLMFollowsForeignKeyPolicy(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateMethodLLM(context.Background(), database.CreateMethodLLMParams{
		MethodID: uuid.New(),
		LlmID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("Table `method_llm`, doesn't follow foreign-key policy.")
	}
}
