package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

type solution_parent_child struct {
	parent_id      uuid.UUID
	child_solution database.CreateSolutionParams
}

func prepareSolution(state config.State) (solution_relations []solution_parent_child, err error) {
	solution_relations = make([]solution_parent_child, 0)
	err = preTestClear(state)
	if err != nil {
		return
	}
	solutions, err := generate_test_sequence()
	if err != nil {
		return
	}
	for i, solution := range solutions {
		solution.input.Generation = sql.NullInt32{Valid: true, Int32: int32(i) + 1}
		for j := i + 1; j < len(solutions); j++ {
			solution_relations = append(solution_relations, solution_parent_child{parent_id: solution.input.ID, child_solution: solutions[j].input})
		}
		_, err = state.DB.CreateSolution(context.Background(), solution.input)
		if err != nil {
			return
		}
	}
	return
}

func TestRelationTableSolutionParentChildInsertion(t *testing.T) {
	state := config.GetContext("test")
	testCases, err := prepareSolution(state)
	if err != nil {
		t.Fatalf("Cannot prepare solutions: %s", err.Error())
	}

	// Insert all parent-child relations.
	for _, testCase := range testCases {
		err = state.DB.CreateParentChild(
			context.Background(),
			database.CreateParentChildParams{
				ParentID: testCase.parent_id,
				ChildID:  testCase.child_solution.ID,
			},
		)
		if err != nil {
			t.Fatalf(
				"Unable to insert into solution_parent_child: %s",
				err,
			)
		}
	}

	// Check whether parents of each solution are valid.
	for _, testCase := range testCases {
		parents, err := state.DB.GetParentsofSolution(
			context.Background(),
			testCase.child_solution.ID,
		)
		if err != nil {
			t.Fatalf(
				"Unable to get parents for solution %s: %s",
				testCase.child_solution.ID,
				err,
			)
		}

		found := false
		for _, parent := range parents {
			if parent.ID == testCase.parent_id {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"Expected solution %s to have parent %s",
				testCase.child_solution.ID,
				testCase.parent_id,
			)
		}
	}

	// Check whether children of each solution are valid.
	for _, testCase := range testCases {
		children, err := state.DB.GetChildrenofSolution(
			context.Background(),
			testCase.parent_id,
		)
		if err != nil {
			t.Fatalf(
				"Unable to get children for solution %s: %s",
				testCase.parent_id,
				err,
			)
		}

		found := false
		for _, child := range children {
			if child.ID == testCase.child_solution.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"Expected solution %s to have child %s",
				testCase.parent_id,
				testCase.child_solution.ID,
			)
		}
	}
}
