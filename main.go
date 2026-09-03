package main

import (
	"anantashahane/BLADE_db/internal/database"
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"

	"database/sql"

	_ "github.com/lib/pq"
)

type state struct {
	db      *database.Queries
	db_path string
}

func main() {
	dbURL := os.Getenv("database_key")
	fmt.Println(dbURL)
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		print("Could not instantiate database.")
		return
	}
	dbQueries := database.New(db)
	app_state := state{db: dbQueries, db_path: dbURL}
	_, err = app_state.db.CreateRun(context.Background(), database.CreateRunParams{
		ID:   uuid.New(),
		Seed: 1,
	})

	if err != nil {
		fmt.Println("Could not create run:", err)
		return
	}

	all_runs, err := app_state.db.GetRuns(context.Background())
	for _, run := range all_runs {
		fmt.Println(run)
	}
}
