package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestBaseTableMessage(t *testing.T) {
	// Cannot insert same message again and again.
	state := config.GetContext("test")
	// preTestClear(state)

	inserted_messages := make([]database.CreateMessageParams, 0)
	message := "hello world"
	digest := sha256.Sum256([]byte(message))
	row := database.CreateMessageParams{ID: uuid.New(), Message: message, Hash: digest[:]}
	id, err := state.DB.CreateMessage(context.Background(), row)
	if err != nil {
		t.Fatalf("Unable to generate first message: %s", err.Error())
	}
	row.ID = id
	inserted_messages = append(inserted_messages, row)

	row = database.CreateMessageParams{ID: uuid.New(), Message: message, Hash: digest[:]}
	id, err = state.DB.CreateMessage(context.Background(), row)
	if err != nil {
		t.Fatal("Insertion of same message returned err; expected behaviour is to provide already existing row's message.id.")
	}
	if id == row.ID {
		t.Fatal("Updating same row, mutated id to new id.")
	}

	// Insertions succeed.
	message1 := "Hello BLADE"
	digest1 := sha256.Sum256([]byte(message1))
	message2 := "Hello PostgreSQL"
	digest2 := sha256.Sum256([]byte(message2))
	message3 := "Hello go."
	digest3 := sha256.Sum256([]byte(message3))
	new_messages := []database.CreateMessageParams{
		{ID: uuid.New(), Message: message1, Hash: digest1[:]},
		{ID: uuid.New(), Message: message2, Hash: digest2[:]},
		{ID: uuid.New(), Message: message3, Hash: digest3[:]},
	}

	for idx, message := range new_messages {
		fmt.Println(idx)
		_, err := state.DB.CreateMessage(context.Background(), message)
		if err != nil {
			t.Fatalf("Unable to insert message %s: %s", message.Message, err.Error())
		}
	}
	inserted_messages = append(inserted_messages, new_messages...)
	// Retrival of all functions succeeds.
	retrived_items, err := state.DB.GetMessages(context.Background())
	fmt.Println(len(inserted_messages), len(retrived_items))
	for index, db_message := range retrived_items {
		fmt.Println(db_message.Message, inserted_messages[index].Message)
		if db_message.ID != inserted_messages[index].ID {
			t.Fatalf("Retrieved ID did not match inserted id %v\t\n%v", retrived_items[index], inserted_messages[index])
		}
		if db_message.Message != inserted_messages[index].Message {
			t.Fatalf("Retrieved message did not match inserted message %v\t\n%v", retrived_items[index], inserted_messages[index])
		}
	}
}
