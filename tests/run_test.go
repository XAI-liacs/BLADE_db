package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func preTestClear(db_config config.State) {
	db_config.DB.ClearRun(context.Background())
	db_config.DB.ClearExperiment(context.Background())
	db_config.DB.ClearTag(context.Background())
	db_config.DB.ClearMessage(context.Background())
	db_config.DB.ClearProblem(context.Background())
	db_config.DB.ClearLLM(context.Background())
	db_config.DB.ClearMethod(context.Background())
	db_config.DB.ClearSolution(context.Background())
}

func TestRun(t *testing.T) {
	state := config.GetContext("test")
	// preTestClear(state)

	// Test insertion of same seed succeeds with different id.
	insertions := make([]database.CreateRunParams, 0, 3)
	inserted_ids := make([]uuid.UUID, 0, 3)

	for range 3 {
		insertions = append(insertions, database.CreateRunParams{ID: uuid.New(), Seed: 1})
		id, err := state.DB.CreateRun(context.Background(), insertions[len(insertions)-1])
		if err != nil {
			t.Fatalf("Insertion error: %s", err.Error())
		}
		inserted_ids = append(inserted_ids, id)
	}

	for i := range 3 {
		if insertions[i].ID != inserted_ids[i] {
			t.Fatalf("ID conflict: inserted %s, got %s", insertions[i].ID, inserted_ids[i])
		}
	}

	// Test Run Get test works.
	for i := range 3 {
		insertions = append(insertions, database.CreateRunParams{ID: uuid.New(), Seed: int32(i + 10)})
		id, err := state.DB.CreateRun(context.Background(), insertions[len(insertions)-1])
		if err != nil {
			t.Fatalf("Insertion error: %s", err.Error())
		}
		inserted_ids = append(inserted_ids, id)
	}

	list, err := state.DB.GetRuns(context.Background())
	if err != nil {
		t.Fatalf("Unable to fetch run....")
	}
	for i, item := range list {
		if item.ID != insertions[i].ID && item.Seed != insertions[i].Seed {
			t.Fatalf("Insertion fetch missmatch: \n\t %v\n\t %v", item, insertions[i])
		}
	}

	random_insertion := insertions[rand.Intn(len(insertions)-1)]
	_, err = state.DB.CreateRun(context.Background(), database.CreateRunParams{ID: random_insertion.ID, Seed: random_insertion.Seed})
	if err == nil {
		t.Fatalf("Database accepted same primary key insertion....")
	}
}
