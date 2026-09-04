package main

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"fmt"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

func main() {
	app_state := config.GetContext("deployment")
	_, err := app_state.DB.CreateRun(context.Background(), database.CreateRunParams{
		ID:   uuid.New(),
		Seed: 1,
	})

	if err != nil {
		fmt.Println("Could not create run:", err)
		return
	}

	all_runs, err := app_state.DB.GetRuns(context.Background())
	for _, run := range all_runs {
		fmt.Println(run)
	}
}
