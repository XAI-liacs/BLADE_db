package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"math/rand"
	"sort"
	"testing"

	"github.com/google/uuid"
)

func TestBaseTableExperiment(t *testing.T) {
	state := config.GetContext("test")
	// preTestClear(state)

	// Test insertion of same seed succeeds with different id.
	insertions := make([]database.CreateExperimentParams, 0, 3)
	inserted_ids := make([]uuid.UUID, 0, 3)

	e1 := getRandomExperiment()
	insertions = append(insertions, e1)
	e2 := e1
	e2.ID = uuid.New()
	insertions = append(insertions, e2)
	e3 := e1
	e3.ID = uuid.New()
	insertions = append(insertions, e3)

	for _, experiment := range insertions {
		id, err := state.DB.CreateExperiment(context.Background(), experiment)
		if err != nil {
			t.Fatalf("Insertion error: %s", err.Error())
		}
		inserted_ids = append(inserted_ids, id)
	}

	for i := range 3 {
		if insertions[i].ID != inserted_ids[i] {
			t.Fatalf("ID conflict: inserted %s, got %s", insertions[i], inserted_ids[i])
		}
	}

	// Test Run Get test works.
	for _ = range 3 {
		experiment := getRandomExperiment()
		insertions = append(insertions, experiment)
		id, err := state.DB.CreateExperiment(context.Background(), experiment)
		if err != nil {
			t.Fatalf("Insertion error: %s", err.Error())
		}
		inserted_ids = append(inserted_ids, id)
	}

	list, err := state.DB.GetExperiments(context.Background())
	if err != nil {
		t.Fatalf("Unable to fetch run....")
	}
	if len(list) == 0 {
		t.Fatalf("Unable nothing imported....")
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID.String() < list[j].ID.String()
	})

	sort.Slice(insertions, func(i, j int) bool {
		return insertions[i].ID.String() < insertions[j].ID.String()
	})

	if len(list) != len(insertions) {
		t.Fatalf("Expected %d insertions, got %d", len(insertions), len(list))
	}

	for i, item := range list {
		if !Equal(insertions[i], item) {
			t.Fatalf("Insertion fetch mismatch:\n\t%v\n\t%v", item, insertions[i])
		}
	}

	random_insertion := insertions[rand.Intn(len(insertions)-1)]
	_, err = state.DB.CreateExperiment(context.Background(), random_insertion)
	if err == nil {
		t.Fatalf("Database accepted same primary key insertion....")
	}
}
