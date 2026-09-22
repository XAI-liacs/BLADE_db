package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
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

	dummy_tag := database.CreateTagParams{ID: uuid.New(), Tag: "Anna"}
	id, err := state.DB.CreateTag(context.Background(), tag)
	if err != nil {
		t.Fatal("Got error while row clash, expected pre-exisitng row's id instead....")
	}
	if id == dummy_tag.ID {
		t.Fatal("Inserting same tag mutated id of older tag.....")
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
