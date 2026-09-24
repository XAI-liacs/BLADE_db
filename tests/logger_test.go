package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLoggerInsertsLog(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateLog(
		context.Background(),
		database.CreateLogParams{
			ID:          uuid.New(),
			Type:        "error",
			Message:     "test logger",
			DatabaseKey: "test",
			CreatedAt:   time.Now(),
		},
	)
	if err != nil {
		t.Fatal("Failed to log.")
	}
}

func TestLoggerDeletesBeforeSpecifiedDate(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateLog(
		context.Background(),
		database.CreateLogParams{
			ID:          uuid.New(),
			Type:        "error",
			Message:     "test logger 1",
			DatabaseKey: "test",
			CreatedAt:   time.Now(),
		},
	)
	err = state.DB.CreateLog(
		context.Background(),
		database.CreateLogParams{
			ID:          uuid.New(),
			Type:        "error",
			Message:     "test logger 2",
			DatabaseKey: "test",
			CreatedAt:   time.Now().Add(time.Hour * -24),
		},
	)
	err = state.DB.CreateLog(
		context.Background(),
		database.CreateLogParams{
			ID:          uuid.New(),
			Type:        "error",
			Message:     "test logger 3",
			DatabaseKey: "test",
			CreatedAt:   time.Now().Add(time.Hour * -36),
		},
	)
	err = state.DB.CreateLog(
		context.Background(),
		database.CreateLogParams{
			ID:          uuid.New(),
			Type:        "error",
			Message:     "test logger 4",
			DatabaseKey: "test",
			CreatedAt:   time.Now().Add(time.Hour * -12),
		},
	)
	err = state.DB.CreateLog(
		context.Background(),
		database.CreateLogParams{
			ID:          uuid.New(),
			Type:        "error",
			Message:     "test logger 5",
			DatabaseKey: "test",
			CreatedAt:   time.Now().Add(time.Hour * -11),
		},
	)
	if err != nil {
		t.Fatal("Failed to log one or more.")
	}

	data, err := state.DB.GetLogs(context.Background())
	if err != nil {
		t.Fatal("Unable to fetch log: " + err.Error())
	}
	if len(data) != 6 {
		t.Fatal("Expected 6 entries, got " + strconv.Itoa(len(data)))
	}
	err = state.DB.ClearLogBefore(context.Background(), time.Now().Add(time.Hour*-12))
	if err != nil {
		t.Fatal("Unable to clear log: " + err.Error())
	}
	data, err = state.DB.GetLogs(context.Background())
	if err != nil {
		t.Fatal("Unable to fetch log: " + err.Error())
	}
	if len(data) != 3 {
		t.Fatal("Expected 3 entries, got " + strconv.Itoa(len(data)))
	}
}
