package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestBaseTableMessage(t *testing.T) {
	// Cannot insert same message again and again.
	state := config.GetContext("test")
	// preTestClear(state)

	inserted_messages := make([]database.CreateMessageParams, 0)

	message := database.CreateMessageParams{ID: uuid.New(), Message: "Hello, world"}
	_, err := state.DB.CreateMessage(context.Background(), message)
	if err != nil {
		t.Fatalf("Unable to generate first message: %s", err.Error())
	}
	inserted_messages = append(inserted_messages, message)
	_, err = state.DB.CreateMessage(context.Background(), message)
	if err == nil {
		t.Fatal("Same messages were imported....")
	}

	// Insertions succeed.
	new_messages := []database.CreateMessageParams{{ID: uuid.New(),
		Message: "Hello BLADE"}, {ID: uuid.New(), Message: "Hello PostgreSQL"},
		{ID: uuid.New(), Message: "Hello go."}}

	for _, message := range new_messages {
		_, err := state.DB.CreateMessage(context.Background(), message)
		if err != nil {
			t.Fatalf("Unable to insert message %s: %s", message.Message, err.Error())
		}
	}
	inserted_messages = append(inserted_messages, new_messages...)
	// Retrival of all functions succeeds.
	retrived_items, err := state.DB.GetMessages(context.Background())
	for index, db_tag := range retrived_items {
		if db_tag.ID != inserted_messages[index].ID {
			t.Fatalf("Retrieved ID did not match inserted id %v\n\n%v", retrived_items, inserted_messages)
		}
		if db_tag.Message != inserted_messages[index].Message {
			t.Fatalf("Retrieved message did not match inserted message %v\n\n%v", retrived_items, inserted_messages)
		}
	}
}
