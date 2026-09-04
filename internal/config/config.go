package config

import (
	"anantashahane/BLADE_db/internal/database"
	"database/sql"
	"errors"
	"os"
)

// State contains the application state required to access the database.
type State struct {
	// DB provides access to the application's database queries.
	DB *database.Queries
	// DB_Path contains the database connection URL used to initialize DB.
	DB_Path string
}

// GetContext initializes the application state for the given environment.
//
// The environment must be either "deployment" or "test". The corresponding
// database connection URL is read from the environment variables
// "database_key" or "test_key".
//
// GetContext panics if the database connection cannot be opened.
func GetContext(environment string) State {
	dbURL := ""
	switch environment {
	case "deployment":
		dbURL = os.Getenv("database_key")
	case "test":
		dbURL = os.Getenv("test_key")
	default:
		panic("Error: no such environement as " + environment + "; use 'deployment' or 'test'")
	}
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		panic(errors.New("Could not instantiate database."))
	}

	dbQueries := database.New(db)

	app_state := State{DB: dbQueries, DB_Path: dbURL}
	return app_state
}
