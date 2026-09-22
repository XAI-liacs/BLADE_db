package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type run_solution struct {
	run_id      uuid.UUID
	seed        int
	solution_id uuid.UUID
	solution    database.CreateSolutionParams
}

func prepareRunAndSolutions(state config.State) (run_solutions []run_solution, err error) {
	run_solutions = make([]run_solution, 0, 3)

	solutions, err := generate_test_sequence()
	if err != nil {
		return run_solutions, err
	}
	for _, solution := range solutions {
		_, err := state.DB.CreateSolution(context.Background(), solution.input)
		if err != nil {
			return run_solutions, errors.New("Unable to populate solution table: " + err.Error())
		}
	}
	for i := range 3 {
		run_id := uuid.New()
		_, err := state.DB.CreateRun(context.Background(), database.CreateRunParams{ID: run_id, Seed: int32(i)})
		if err != nil {
			return run_solutions, errors.New("Unable to populate run table: " + err.Error())
		}
		if i != 2 {
			solution_id := solutions[i].input.ID
			run_solutions = append(run_solutions, run_solution{
				run_id:      run_id,
				seed:        i,
				solution_id: solution_id,
				solution:    solutions[i].input,
			})
		} else {
			for _, solution := range solutions {
				run_solutions = append(run_solutions, run_solution{
					run_id:      run_id,
					seed:        i,
					solution_id: solution.input.ID,
					solution:    solution.input,
				})
			}
		}
	}
	return run_solutions, nil

}

func TestRelationTableRunSolutionInsertion(t *testing.T) {
	state := config.GetContext("test")

	run_solutions, err := prepareRunAndSolutions(state)
	if err != nil {
		t.Fatalf("Unable to prepare run and solution tables: %s", err.Error())
	}

	for _, runSolution := range run_solutions {
		row := database.CreateRunSolutionParams{
			RunID:      runSolution.run_id,
			SolutionID: runSolution.solution_id,
		}

		err = state.DB.CreateRunSolution(context.Background(), row)
		if err != nil {
			t.Fatalf("Unable to add %v to run solution table: %s", row, err.Error())
		}
	}

	// Group expected solutions by run.
	expected := make(map[uuid.UUID][]run_solution)

	for _, rs := range run_solutions {
		expected[rs.run_id] = append(expected[rs.run_id], rs)
	}

	// Test validity (Read).
	for runID, expectedSolutions := range expected {
		fetchedSolutions, err := state.DB.GetSolutionsForRun(
			context.Background(),
			runID,
		)
		if err != nil {
			t.Fatalf(
				"Unable to fetch solutions for run id %s: %s",
				runID,
				err.Error(),
			)
		}

		if len(fetchedSolutions) != len(expectedSolutions) {
			t.Fatalf(
				"Run %s: expected %d solutions, got %d",
				runID,
				len(expectedSolutions),
				len(fetchedSolutions),
			)
		}

		expectedByID := make(map[uuid.UUID]run_solution)

		for _, rs := range expectedSolutions {
			expectedByID[rs.solution_id] = rs
		}

		for _, fetchedSolution := range fetchedSolutions {
			expectedSolution, ok := expectedByID[fetchedSolution.ID]
			if !ok {
				t.Fatalf(
					"Run %s: unexpected solution ID %s",
					runID,
					fetchedSolution.ID,
				)
			}

			if !optional_string_equality(
				fetchedSolution.Code,
				expectedSolution.solution.Code,
			) {
				t.Fatalf(
					"Code mismatch for solution %s: %v | %v",
					fetchedSolution.ID,
					fetchedSolution.Code,
					expectedSolution.solution.Code,
				)
			}

			if !reflect.DeepEqual(
				fetchedSolution.Generation,
				expectedSolution.solution.Generation,
			) {
				t.Fatalf(
					"Generation mismatch for solution %s: %v | %v",
					fetchedSolution.ID,
					fetchedSolution.Generation,
					expectedSolution.solution.Generation,
				)
			}

			if !isEqualNullableJsonObject(
				fetchedSolution.Fitness,
				expectedSolution.solution.Fitness,
			) {
				t.Fatalf(
					"Fitness mismatch for solution %s: %v | %v",
					fetchedSolution.ID,
					fetchedSolution.Fitness,
					expectedSolution.solution.Fitness,
				)
			}

			if !isEqualNullableJsonObject(
				fetchedSolution.Metadata,
				expectedSolution.solution.Metadata,
			) {
				t.Fatalf(
					"Metadata mismatch for solution %s: %v | %v",
					fetchedSolution.ID,
					fetchedSolution.Metadata,
					expectedSolution.solution.Metadata,
				)
			}

			if !optional_string_equality(
				fetchedSolution.Name,
				expectedSolution.solution.Name,
			) {
				t.Fatalf(
					"Name mismatch for solution %s: %v | %v",
					fetchedSolution.ID,
					fetchedSolution.Name,
					expectedSolution.solution.Name,
				)
			}
		}
	}
}

func TestRelationTableRunSolutionRejectsUnknownIDs(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateRunSolution(context.Background(), database.CreateRunSolutionParams{
		RunID:      uuid.New(),
		SolutionID: uuid.New(),
	})
	if err == nil {
		t.Fatalf("DB accepted unknown keys.")
	}
}
