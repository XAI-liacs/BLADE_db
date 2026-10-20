package main

import (
	"XAI-liacs/BLADE_db/internal/config"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	state := config.GetContext("deployment")
	preTestClear(state)

	mux := http.NewServeMux()

	mux.Handle("/frontend/",
		http.StripPrefix("/frontend/",
			http.FileServer(http.Dir("frontend")),
		),
	) //Serve scripts and styles as a filesystem downloadables.

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		homeHandler(w, r)
	})

	mux.HandleFunc("/html/experiments", HandleHtmlExperiment)
	mux.HandleFunc("/html/run_data", HandleHtmlRunData)
	mux.HandleFunc("/html/submit", submitHandler) // Used to upload an experiment file to the server.

	mux.HandleFunc("/health", healthHandler) // Health check, returns ok, and supported blade verison as response.

	mux.HandleFunc("/signup", signUpHandler) // Used to Signup as a user.
	mux.HandleFunc("/login", loginHandler)   // Used to Login as a user.

	mux.HandleFunc("api/runs", getRunsHandler)                       // Gets runs for a given experiment = /runs?experiment_id=<experiment_id>
	mux.HandleFunc("api/solutions", getSolutionsLogHandler)          // Gets solutions for a given run = /solutions?run_id=<run_id>
	mux.HandleFunc("api/experiments", getExperimentsHandler)         // Gets experiments uploaded by the user #TODO: add JWT support to show only the users' information.
	mux.HandleFunc("api/conversationlog", getConversationLogHandler) // Gets conversation log for the server.

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
