package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

func prepareRunTable(state config.State) (run_ids []uuid.UUID, err error) {
	run_ids = make([]uuid.UUID, 0)
	for i := range 3 {
		row := database.CreateRunParams{ID: uuid.New(), Seed: int32(i)}
		id, err := state.DB.CreateRun(context.Background(), row)
		if err != nil {
			return run_ids, errors.New("Unable to generate run, error: " + err.Error())
		}
		run_ids = append(run_ids, id)
	}
	return run_ids, nil
}

func prepareRunDescriptorRequiredTables(state config.State) (run_descriptors []database.CreateRunDescriptorParams, err error) {
	preTestClear(state)
	run_descriptors = make([]database.CreateRunDescriptorParams, 0)
	run_ids, err := prepareRunTable(state)
	if err != nil {
		return
	}
	_, err = prepareTagAndProblem(state)
	if err != nil {
		return
	}

	problem_obj, err := state.DB.GetProblems(context.Background())
	if err != nil {
		return
	}
	_, err = prepareMethod(state)
	method_obj, err := state.DB.GetMethods(context.Background())
	if err != nil {
		return
	}

	for run_index, run_id := range run_ids {
		for i := 0; i <= run_index; i++ {
			problem_id := problem_obj[i].ID
			method_id := method_obj[rand.IntN(len(method_obj))].ID
			run_descriptors = append(run_descriptors, database.CreateRunDescriptorParams{ProblemID: problem_id, MethodID: method_id, RunID: run_id})
		}
	}
	return
}

func TestRelationTableRunDescriptorInsertion(t *testing.T) {
	state := config.GetContext("test")
	run_descriptors, err := prepareRunDescriptorRequiredTables(state)
	if err != nil {
		t.Fatal(err.Error())
	}
	// Insertions.
	for _, descriptor := range run_descriptors {
		err := state.DB.CreateRunDescriptor(context.Background(), descriptor)
		if err != nil {
			t.Fatalf("Unable to insert %v, into run_descriptor table: %s", descriptor, err.Error())
		}
	}

	// Reads.
	for _, descriptor := range run_descriptors {
		found := false
		fetched_descriptors, err := state.DB.GetMethodAndProblemForRun(context.Background(), descriptor.RunID)
		if err != nil {
			t.Fatalf("Unable to fetch run descriptor for run_id: %v, error: %s", descriptor.RunID, err.Error())
		}
		for _, fetched_row := range fetched_descriptors {
			if fetched_row.MethodID == descriptor.MethodID && fetched_row.ProblemID == descriptor.ProblemID {
				found = true
			}
		}
		if !found {
			t.Fatalf("Cannot find %v in the fetched results...\n\t%v", descriptor, fetched_descriptors)
		}
	}
}

func TestRelationTableRunDescriptorFollowsForeignKeyPolicy(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateRunDescriptor(context.Background(), database.CreateRunDescriptorParams{
		RunID:     uuid.New(),
		ProblemID: uuid.New(),
		MethodID:  uuid.New(),
	})
	if err == nil {
		t.Fatal("Table run_descriptor ignored foreign key policy...")
	}
}
