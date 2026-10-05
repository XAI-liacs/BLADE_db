package main

import (
	"XAI-liacs/BLADE_db/authentication"
	filehandlers "XAI-liacs/BLADE_db/file_handlers"
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxUploadSize = 16 << 30 // 2 GB.

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)

		// Prevent ZIP path traversal.
		if !strings.HasPrefix(
			filepath.Clean(target)+string(os.PathSeparator),
			filepath.Clean(dest)+string(os.PathSeparator),
		) {
			return fmt.Errorf("invalid zip path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}

		in, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(
			target,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0600,
		)
		if err != nil {
			in.Close()
			return err
		}

		_, copyErr := io.Copy(out, in)

		closeOutErr := out.Close()
		closeInErr := in.Close()

		if copyErr != nil {
			return copyErr
		}
		if closeOutErr != nil {
			return closeOutErr
		}
		if closeInErr != nil {
			return closeInErr
		}
	}

	return nil
}

type HealthIndicator struct {
	Health  string
	Version string
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
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

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	type SignUpUser struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	state := config.GetContext("deployment")
	tx, err := state.DB_pointer.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Unable to start transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	state.DB = state.DB.WithTx(tx)

	user_data := SignUpUser{}

	err = json.NewDecoder(r.Body).Decode(&user_data)

	if err != nil {
		http.Error(w, "invalid JSON"+err.Error(), http.StatusBadRequest)
		return
	}

	username_regex := regexp.MustCompile(`^[a-zA-Z0-9._-]{4,32}$`)
	if !username_regex.MatchString(user_data.Name) {
		http.Error(w, "Invalid User name.", http.StatusBadRequest)
		return
	}

	if len(user_data.Password) < 10 {
		http.Error(w, "Password too short.", http.StatusBadRequest)
		return
	}

	_, err = state.DB.GetUserNamed(r.Context(), user_data.Name)
	if err == nil {
		http.Error(w, "User already exists.", http.StatusBadRequest)
		return
	}
	pass_hash := sha256.Sum256([]byte(user_data.Password))

	user_id, err := state.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		Name:         user_data.Name,
		PasswordHash: pass_hash[:],
		IsAdmin:      false,
	})
	if err != nil {
		http.Error(w, "Unable to create user, "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User ID: "+user_id.String())
	tx.Commit()
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginUser struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	state := config.GetContext("deployment")

	user_data := LoginUser{}

	err := json.NewDecoder(r.Body).Decode(&user_data)

	if err != nil {
		http.Error(w, "invalid JSON"+err.Error(), http.StatusBadRequest)
		return
	}

	username_regex := regexp.MustCompile(`^[a-zA-Z0-9._-]{4,32}$`)
	if !username_regex.MatchString(user_data.Name) {
		http.Error(w, "Invalid User name.", http.StatusBadRequest)
		return
	}

	if len(user_data.Password) < 10 {
		http.Error(w, "Password too short.", http.StatusBadRequest)
		return
	}

	db_user, err := state.DB.GetUserNamed(r.Context(), user_data.Name)
	if err != nil {
		http.Error(w, "User not signed up.", http.StatusBadRequest)
		return
	}
	pass_hash := sha256.Sum256([]byte(user_data.Password))

	if pass_hash != [32]byte(db_user.PasswordHash) {
		http.Error(w, "Incorrect password.", http.StatusBadRequest)
		return
	}
	expiration_time := (60 * time.Hour * 24)
	jwt_token, err := authentication.MakeJWT(db_user.ID, expiration_time)
	if err != nil {
		http.Error(w, "Unable to create web token, "+err.Error(), http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    jwt_token,
		Path:     "/dashboard",
		HttpOnly: true,  //Security against js.
		Secure:   false, // HTTPS only #TODO: set true when deploying.
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(expiration_time),
	})

	w.WriteHeader(http.StatusOK)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Check who is uploading the file.
	token, err := authentication.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorised upload; sign in to upload", http.StatusBadRequest)
		return
	}
	user_id, err := authentication.ValidateJWT(token)
	if err != nil {
		http.Error(w, "Invalid token; sign in to upload", http.StatusBadRequest)
		return
	}
	// Protect against accidentally huge uploads.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	defer r.Body.Close()

	id := uuid.New()
	filename := fmt.Sprintf("blade_db_%s", id.String()[:8])
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
	errs := filehandlers.ImportAllExperimentUnder(unzipped_path, "deployment", user_id)
	if len(errs) != 0 {
		http.Error(w, "Failed to save to db.", http.StatusInternalServerError)
		return
	} else {
		// Clean up directory on success, let the directory be, if failure.
		temp_dir := filepath.Dir(path)
		os.RemoveAll(temp_dir)
	}
	fmt.Printf("Received upload: %s\n", path)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "uploaded successfully.\n")

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

func preTestClear(dbConfig config.State) error {
	fmt.Println("\t=== Clearing all tables ===")
	ctx := context.Background()
	// Clear Relation Tables First.....
	if err := dbConfig.DB.ClearUserExperiments(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearRunDescriptor(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearConversationLog(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearMethodLLM(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearTagProblem(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearExperimentRun(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearParentChild(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearRunSolution(ctx); err != nil {
		return err
	}

	// Clear Base Tables Later.....

	if err := dbConfig.DB.ClearUsers(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearLogBefore(ctx, time.Now()); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearTag(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearRun(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearLLM(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearProblem(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearMessage(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearExperiment(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearMethod(ctx); err != nil {
		return err
	}
	if err := dbConfig.DB.ClearSolution(ctx); err != nil {
		return err
	}
	return nil
}
