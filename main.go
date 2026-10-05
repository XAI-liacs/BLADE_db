package main

import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	// state := config.GetContext("deployment")
	// preTestClear(state)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler) // Health check, returns ok, and supported blade verison as response.

	mux.HandleFunc("/signup", signUpHandler) // Used to Sign up as a user.
	mux.HandleFunc("/submit", submitHandler) // Used to upload an experiment file to the server.

	mux.HandleFunc("/runs", getRunsHandler)                       // Gets runs for a given experiment = /runs?experiment_id=<experiment_id>
	mux.HandleFunc("/solutions", getSolutionsLogHandler)          // Gets solutions for a given run = /solutions?run_id=<run_id>
	mux.HandleFunc("/experiments", getExperimentsHandler)         // Gets experiments uploaded by the user #TODO: add JWT support to show only the users' information.
	mux.HandleFunc("/conversationlog", getConversationLogHandler) // Gets conversation log for the server.

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	fmt.Println("Server listening on :8080")
	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		panic(err)
	}
}
