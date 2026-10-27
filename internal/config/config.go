package config

import (
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
)

// State contains the application state required to access the database.
type State struct {
	// DB provides access to the application's database queries.
	DB *database.Queries
	// Database connector:
	DB_pointer *sql.DB
	// DB_Path contains the database connection URL used to initialize DB.
	DB_Path string
	// Transaction pointer, to pass transaction to commit/revert:
	TX *sql.Tx
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

	app_state := State{DB: dbQueries, DB_pointer: db, DB_Path: dbURL}
	return app_state
}

// Retrieves the database state for the specified backend and starts a new database transaction.
//
// ## Params
//
// `context: context.Context` Context used to control the lifetime of the transaction.
//
// ## Returns:
//
// `state: config.State`: Returns the database state configured to use the newly started transaction.
//
// `error`: Returns an error if the transaction could not be started; otherwise, returns nil.
func (state *State) GetTransactionalState(ctx context.Context) (*State, error) {
	tx, err := state.DB_pointer.BeginTx(ctx, nil)
	if err != nil {
		return &State{}, fmt.Errorf("unable to start transaction: %w", err)
	}

	db_connection := state.DB.WithTx(tx)
	return &State{
		DB_Path:    state.DB_Path,
		DB:         db_connection,
		DB_pointer: state.DB_pointer,
		TX:         tx,
	}, nil
}
