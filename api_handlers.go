package main

import (
	filehandlers "XAI-liacs/BLADE_db/file_handlers"
	"XAI-liacs/BLADE_db/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type HealthIndicator struct {
		Health  string
		Version string
	}
	w.WriteHeader(http.StatusOK)

	health := HealthIndicator{
		Health:  "Ok",
		Version: "1.4.7",
	}

	jsonData, err := json.Marshal(health)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Println(r.RemoteAddr, r.Pattern, r.Body)
	w.Write(jsonData)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Check who is uploading the file.
	user, err := authenticate(r, "html")
	if err != nil {
		http.Error(w, "Unauthorised access error: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Protect against accidentally huge uploads.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	defer r.Body.Close()

	filename := fmt.Sprintf("blade_db_%s", uuid.New().String()[:8])
	// Create the file in the system temp directory.
	path := filepath.Join(os.TempDir(), filename)
	if err := os.Mkdir(path, 0700); err != nil {
		http.Error(w, "failed to create temporary file"+path+" Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	path = filepath.Join(path, "download.zip")
	file, err := os.Create(path)
	if err != nil {
		http.Error(w, "failed to create temporary file"+path+" Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// Stream the upload directly to disk.
	_, err = io.Copy(file, r.Body)
	if err != nil {
		// Remove a potentially incomplete upload.
		http.Error(w, "failed to save upload", http.StatusInternalServerError)
		return
	}
	unzipped_path := filepath.Dir(path)

	unzip(path, unzipped_path)
	errs := filehandlers.ImportAllExperimentUnder(unzipped_path, "deployment", user)
	if len(errs) != 0 {
		var builder strings.Builder

		for _, error_info := range errs {
			builder.WriteString("\t")
			builder.WriteString(error_info.Error())
			builder.WriteString("\n")
		}
		http.Error(w, "Failed to save to db: \n"+builder.String(), http.StatusInternalServerError)
		return
	} else {
		// Clean up directory on success, let the directory be, if failure.
		temp_dir := filepath.Dir(path)
		os.RemoveAll(temp_dir)
	}
	log.Printf("Received file from %s: %s\n", user.Name, path)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Experiments Uploaded successfully.\n")
}

func getExperimentsHandler(w http.ResponseWriter, r *http.Request) {
	state := config.GetContext("deployment")
	experiments, err := state.DB.GetExperiments(context.Background())
	if err != nil {
		http.Error(w, "Unable to fetch experiments "+err.Error(), http.StatusInternalServerError)
		return
	}

	bin_data, err := json.Marshal(experiments)
	if err != nil {
		http.Error(w, "Unable to marshal experiments "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bin_data)
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Returned experiments..")
}

func getRunsHandler(w http.ResponseWriter, r *http.Request) {
	state := config.GetContext("deployment")

	experimentID, err := uuid.Parse(r.URL.Query().Get("experiment_id"))
	if err != nil {
		http.Error(w, "Unable to fetch parse experiment ID "+err.Error(), http.StatusBadRequest)
		return
	}

	runs, err := state.DB.GetRunsForExperiment(context.Background(), experimentID)
	if err != nil {
		http.Error(w, "Unable to fetch runs for experiment "+err.Error(), http.StatusInternalServerError)
		return
	}

	bin_data, err := json.Marshal(runs)
	if err != nil {
		http.Error(w, "Unable to marshal runs "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bin_data)
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Returned runs..")
}

func getConversationLogHandler(w http.ResponseWriter, r *http.Request) {
	state := config.GetContext("deployment")

	run_id, err := uuid.Parse(r.URL.Query().Get("run_id"))
	if err != nil {
		http.Error(w, "Unable to fetch parse run ID "+err.Error(), http.StatusBadRequest)
		return
	}

	runs, err := state.DB.GetConversationLog(context.Background(), run_id)
	if err != nil {
		http.Error(w, "Unable to fetch conversation log for run "+err.Error(), http.StatusInternalServerError)
		return
	}

	bin_data, err := json.MarshalIndent(runs, "", "   ")
	if err != nil {
		http.Error(w, "Unable to marshal conversation log "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bin_data)
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Returned converation logs..")
}

func getSolutionsLogHandler(w http.ResponseWriter, r *http.Request) {
	state := config.GetContext("deployment")

	run_id, err := uuid.Parse(r.URL.Query().Get("run_id"))
	if err != nil {
		http.Error(w, "Unable to fetch parse run ID "+err.Error(), http.StatusBadRequest)
		return
	}

	solutions, err := state.DB.GetSolutionsForRun(context.Background(), run_id)
	if err != nil {
		http.Error(w, "Unable to fetch solutions for run "+err.Error(), http.StatusInternalServerError)
		return
	}

	bin_data, err := json.MarshalIndent(solutions, "", "   ")
	if err != nil {
		http.Error(w, "Unable to marshal solutions "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bin_data)
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Returned solutions..")
}
