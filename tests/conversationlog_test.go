package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type conversationLog struct {
	run        database.CreateRunParams
	message    database.CreateMessageParams
	method     *database.CreateMethodParams
	llm        *database.CreateLLMParams
	created_at time.Time
}

var mockPrompts = []string{
	"Summarize this document.",
	"Explain this algorithm.",
	"Find the bug in this code.",
	"Optimize this SQL query.",
	"Write a Go function for sorting.",
	"Convert this JSON to CSV.",
	"What does this error mean?",
	"Suggest a better data structure.",
	"Generate unit tests for this function.",
	"Explain this API response.",
	"Refactor this code.",
	"Compare these two approaches.",
	"Generate a regex for email validation.",
	"How can I reduce memory usage?",
	"Write a database migration.",
	"Explain the time complexity.",
	"Generate sample test data.",
	"Fix this concurrency issue.",
	"Suggest an indexing strategy.",
	"Review this implementation.",
}

var mockResponses = []string{
	"Here is a concise summary of the document.",
	"The algorithm processes each item and updates the current state.",
	"The issue is caused by an incorrect boundary condition.",
	"Add an index on the columns used in the WHERE clause.",
	"Here is a Go implementation using a simple sorting approach.",
	"The JSON fields can be mapped directly to CSV columns.",
	"The error indicates that the requested resource was not found.",
	"A hash map provides faster lookups for this use case.",
	"Here are unit tests covering the main and edge cases.",
	"The response contains the requested data and metadata.",
	"I would simplify the control flow and remove duplicated logic.",
	"The first approach is simpler, while the second scales better.",
	"Use a pattern that validates the local and domain portions.",
	"Reduce allocations and reuse buffers where possible.",
	"Create the new table and add the required constraints.",
	"The time complexity is O(n log n) in the average case.",
	"Here is a small dataset covering normal and edge cases.",
	"The race occurs because multiple goroutines access shared state.",
	"An index on the filtered and frequently joined columns should help.",
	"The implementation is mostly sound, but error handling can be improved.",
}

func simulateConversation(length int) (conversation []string) {
	conversation = make([]string, 0)
	for _ = range length {
		index := rand.IntN(len(mockPrompts))
		agent := mockPrompts[index]
		llm := mockResponses[index]
		conversation = append(conversation, agent)
		conversation = append(conversation, llm)
	}
	return
}

func createConversationLogRow(run database.CreateRunParams,
	method *database.CreateMethodParams,
	llms []database.CreateLLMParams,
	started_at time.Time) (conversation []conversationLog) {
	messages := simulateConversation(50)
	for index, message := range messages {
		hash := sha256.Sum256([]byte(message))
		message_obj := database.CreateMessageParams{
			ID:      uuid.New(),
			Message: message,
			Hash:    hash[:],
		}
		if index%2 == 0 {
			conversation_row := conversationLog{
				run:        run,
				message:    message_obj,
				method:     nil,
				llm:        &llms[rand.IntN(len(llms))],
				created_at: started_at.Add(time.Duration(index*5) * time.Second),
			}
			conversation = append(conversation, conversation_row)
		} else {
			conversation_row := conversationLog{
				run:        run,
				message:    message_obj,
				method:     method,
				llm:        nil,
				created_at: started_at.Add(time.Duration(index*5) * time.Second),
			}
			conversation = append(conversation, conversation_row)
		}
	}
	return conversation
}

func createConversationLogInstance(run_number int32) []conversationLog {
	methods := make([]string, 2)
	// log := make([]conversationLog, 0)
	methods[0] = "LLaMEA"
	methods[1] = "MCTS-AHD"

	run := database.CreateRunParams{
		ID:   uuid.New(),
		Seed: int32(run_number),
	}
	method_name := methods[rand.IntN(len(methods))]
	config_data := make(map[string]any)
	config_json, _ := json.Marshal(config_data)
	method := database.CreateMethodParams{
		ID:     uuid.New(),
		Name:   method_name,
		Source: "https://www.method.lib/" + method_name,
		Config: config_json,
	}

	config_data["temperature"] = 0.8
	config_json, _ = json.Marshal(config_data)
	llms := make([]database.CreateLLMParams, 0) //model-multi-llm
	for llm_index := range run_number + 1 {
		llms = append(llms, database.CreateLLMParams{
			ID:       uuid.New(),
			Model:    fmt.Sprintf("gemma_%v", llm_index),
			Hardware: pqtype.NullRawMessage{RawMessage: config_json, Valid: false},
			Config:   config_json,
		})
	}
	return createConversationLogRow(run, &method, llms, time.Now())
}

func TestRelationTableConversationLogInsertion(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)

	for index := range 3 {

		unique_llms := make(map[uuid.UUID]*database.CreateLLMParams)
		unique_methods := make(map[uuid.UUID]*database.CreateMethodParams)

		conv_log := createConversationLogInstance(int32(index))
		run := conv_log[0].run

		run_id, err := state.DB.CreateRun(context.Background(), run)
		if err != nil {
			t.Fatal(err)
		}

		for _, message := range conv_log {
			if message.llm != nil {
				unique_llms[message.llm.ID] = message.llm
			} else {
				unique_methods[message.method.ID] = message.method
			}
		}
		db_llm_ids := make(map[uuid.UUID]uuid.UUID)    //DB ids.
		db_method_ids := make(map[uuid.UUID]uuid.UUID) //DB ids.

		for llm_id, llm := range unique_llms {
			id, err := state.DB.CreateLLM(context.Background(), *llm)
			if err != nil {
				t.Fatal(err.Error())
			}
			db_llm_ids[llm_id] = id
		}

		for method_id, method := range unique_methods {
			id, err := state.DB.CreateMethod(context.Background(), *method)
			if err != nil {
				t.Fatal(err.Error())
			}
			db_method_ids[method_id] = id
		}
		// Prepare message table and generate conoversationLog entries.
		conversation_log_rows := make([]database.CreateConversationLogParams, 0)
		for _, message := range conv_log {
			message_id, err := state.DB.CreateMessage(context.Background(), message.message)
			if err != nil {
				t.Fatal(err.Error())
			}
			if message.method == nil {
				row := database.CreateConversationLogParams{
					RunID:     run_id,
					MessageID: message_id,
					MethodID:  uuid.NullUUID{UUID: message_id, Valid: false},
					LlmID:     uuid.NullUUID{UUID: db_llm_ids[message.llm.ID], Valid: true},
					CreatedAt: message.created_at,
				}

				if err = state.DB.CreateConversationLog(context.Background(), row); err != nil {
					t.Fatal(err.Error())
				}
				conversation_log_rows = append(conversation_log_rows, row)
			} else {
				row := database.CreateConversationLogParams{
					RunID:     run_id,
					MessageID: message_id,
					MethodID:  uuid.NullUUID{UUID: db_method_ids[message.method.ID], Valid: true},
					LlmID:     uuid.NullUUID{UUID: db_llm_ids[message_id], Valid: false},
					CreatedAt: message.created_at}

				if err = state.DB.CreateConversationLog(context.Background(), row); err != nil {
					t.Fatal(err.Error())
				}
				conversation_log_rows = append(conversation_log_rows, row)
			}
		}
		data, err := state.DB.GetConversationLog(context.Background(), run_id)
		if err != nil {
			t.Fatalf("Unable to fetch conversation for run_id (%v): %s", run_id, err.Error())
		}
		if len(data) != len(conversation_log_rows) {
			t.Fatal("Insertion mis-match.")
		}
	}
}

func TestRelationTableConversationLogRejectsForgedIDs(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateConversationLog(context.Background(), database.CreateConversationLogParams{
		RunID:     uuid.New(),
		MessageID: uuid.New(),
		MethodID:  uuid.NullUUID{UUID: uuid.New(), Valid: false},
		LlmID:     uuid.NullUUID{UUID: uuid.New(), Valid: true},
		CreatedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("Random row imported in ConversationLog table....")
	}
}
