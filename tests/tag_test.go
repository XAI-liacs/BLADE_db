package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestBaseTableTag(t *testing.T) {
	// Cannot insert same tags again and again.
	state := config.GetContext("test")
	state.DB.ClearTag(context.Background())

	inserted_tags := make([]database.CreateTagParams, 0)

	tag := database.CreateTagParams{ID: uuid.New(), Tag: "Anna"}
	_, err := state.DB.CreateTag(context.Background(), tag)
	if err != nil {
		t.Fatalf("Unable to generate first tag: %s", err.Error())
	}
	inserted_tags = append(inserted_tags, tag)
	_, err = state.DB.CreateTag(context.Background(), tag)
	if err == nil {
		t.Fatal("Same named tags were imported....")
	}

	// Insertions succeed.
	new_tags := []database.CreateTagParams{{ID: uuid.New(),
		Tag: "Optimisation"}, {ID: uuid.New(), Tag: "Bayesian Optimisation"},
		{ID: uuid.New(), Tag: "Black Box Optimisation"}}

	for _, tag := range new_tags {
		_, err := state.DB.CreateTag(context.Background(), tag)
		if err != nil {
			t.Fatalf("Unable to generate tag %s: %s", tag, err.Error())
		}
	}
	inserted_tags = append(inserted_tags, new_tags...)
	// Retrival of all functions succeeds.
	retrived_items, err := state.DB.GetTags(context.Background())
	for index, db_tag := range retrived_items {
		if db_tag.ID != inserted_tags[index].ID {
			t.Fatalf("Retrieved ID did not match insertied id %v\n\n%v", retrived_items, inserted_tags)
		}
		if db_tag.Tag != inserted_tags[index].Tag {
			t.Fatalf("Retrieved tag did not match insertied id %v\n\n%v", retrived_items, inserted_tags)
		}
	}
}
