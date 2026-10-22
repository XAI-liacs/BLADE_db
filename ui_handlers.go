package main

import (
	"XAI-liacs/BLADE_db/authentication"
	"XAI-liacs/BLADE_db/components"
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"crypto/sha256"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/google/uuid"
)

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	state := config.GetContext("deployment")
	tx, err := state.DB_pointer.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Unable to start transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	state.DB = state.DB.WithTx(tx)

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	username_regex := regexp.MustCompile(`^[a-zA-Z0-9._-]{4,32}$`)
	if !username_regex.MatchString(username) {
		http.Error(w, "Invalid User name.", http.StatusBadRequest)
		return
	}

	if len(password) < 10 {
		http.Error(w, "Password too short, need 10 or more characters.", http.StatusBadRequest)
		return
	}

	_, err = state.DB.GetUserNamed(r.Context(), username)
	if err == nil {
		http.Error(w, "User already exists.", http.StatusBadRequest)
		return
	}
	pass_hash := sha256.Sum256([]byte(password))

	_, err = state.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		Name:         username,
		PasswordHash: pass_hash[:],
		IsAdmin:      false,
	})
	if err != nil {
		http.Error(w, "Unable to create user, "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "Unable to commit user creation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	state := config.GetContext("deployment")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	db_user, err := state.DB.GetUserNamed(r.Context(), username)
	if err != nil {
		http.Error(w, "User not signed up.", http.StatusBadRequest)
		return
	}
	pass_hash := sha256.Sum256([]byte(password))

	if pass_hash != [32]byte(db_user.PasswordHash) {
		http.Error(w, "Incorrect password.", http.StatusBadRequest)
		return
	}
	expiration_time := (60 * time.Hour * 24)
	jwt_token, err := authentication.MakeJWT(db_user.ID, expiration_time)
	if err != nil {
		http.Error(w, "Unable to create web token, "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    jwt_token,
		Path:     "/",
		HttpOnly: true,  //Security against js.
		Secure:   false, // HTTPS only #TODO: set true when deploying.
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(expiration_time),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := authenticate(r, "html")
	if err != nil {
		fmt.Println(err.Error())
		tmpl, err := template.ParseFiles("frontend/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}
	tmpl, err := template.ParseFiles("frontend/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func HandleHtmlExperiment(w http.ResponseWriter, r *http.Request) {
	component := components.EmptyExperiment()
	user, err := authenticate(r, "html")
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	state, err := getDBConnection("deployment", r.Context(), false)
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	experiments, err := state.DB.GetExperimentsForUser(context.Background(), user.ID)
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	if len(experiments) == 0 {
		component.Render(r.Context(), w)
	}

	experiment_list := components.ExperimentList(experiments)
	experiment_list.Render(r.Context(), w)
}

func HandleRunCells(w http.ResponseWriter, r *http.Request) {
	user, err := authenticate(r, "html")
	if err != nil {
		log.Println("Got unauthorised access to /html/run" + err.Error())
		http.Error(w, "Unauthorised access.", http.StatusUnauthorized)
		return
	}

	experiment_id, err := uuid.Parse(r.URL.Query().Get("experiment_id"))
	log.Println("Request URL:", r.URL.String())
	if err != nil {
		log.Println("Query error in /html/run" + err.Error())
		http.Error(w, "Bad access error, expected /html/runs?experiment_id=<some uuid>", http.StatusBadRequest)
		return
	}

	db, err := getDBConnection("deployment", r.Context(), false)
	if err != nil {
		log.Println("Unable to instantiate DB connection in /html/run: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ownership_map, err := db.DB.GetUsersAndVisibilityOfExperiment(r.Context(), experiment_id)
	if err != nil {
		log.Println("Got error loading viewability in /html/run: ", err.Error())
		http.Error(w, "DB error.", http.StatusInternalServerError)
		return
	}
	if !(slices.Contains(ownership_map.UserIds, user.ID) || ownership_map.IsPublic) {
		http.Error(w, "Unauthorised access", http.StatusInternalServerError)
		return
	}

	runs, err := db.DB.GetRunsForExperiment(r.Context(), experiment_id)
	if err != nil {
		log.Println("Got error loading runs for expreriment "+experiment_id.String()+" in /html/run: ", err.Error())
		http.Error(w, "DB access error.", http.StatusInternalServerError)
		return
	}
	runCells := components.RunCells(runs, experiment_id, slices.Contains(ownership_map.UserIds, user.ID), ownership_map.IsPublic)
	runCells.Render(r.Context(), w)
}

func HandleHtmlRunData(w http.ResponseWriter, r *http.Request) {
	component := components.EmptyRightPanel("No runs selected.")

	token, err := authentication.GetCookie(r.Header)
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	_, err = authentication.ValidateJWT(token)
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	state := config.GetContext("deployment")

	runID, err := uuid.Parse(r.URL.Query().Get("run_id"))
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	solutions, err := state.DB.GetSolutionsForRun(r.Context(), runID)
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	conversationLog, err := state.DB.GetConversationLog(r.Context(), runID)
	if err != nil {
		conversationLog = []database.GetConversationLogRow{}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	components.RunData(solutions, conversationLog).Render(r.Context(), w)
}

func HandleHtmlVisibilityToggle(w http.ResponseWriter, r *http.Request) {
	_, err := authenticate(r, "html")
	if err != nil {
		log.Println("Got unauthorised access to /html/togle_visibility" + err.Error())
		http.Error(w, "Unauthorised access.", http.StatusUnauthorized)
		return
	}
	experiment_id, err := uuid.Parse(r.URL.Query().Get("experiment_id"))
	log.Println("Request URL:", r.URL.String())
	if err != nil {
		log.Println("Query error in /html/togle_visibility" + err.Error())
		http.Error(w, "Bad access error, expected /html/toggle_visibility?experiment_id=<some uuid>", http.StatusBadRequest)
		return
	}

	db, err := getDBConnection("deployment", r.Context(), true)
	if err != nil {
		log.Println("Unable to instantiate DB connection in /html/toggle_visibility: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.TX.Commit()

	visibility, err := db.DB.ToggleExperimentVisibility(r.Context(), experiment_id)
	if err != nil {
		db.TX.Rollback()
		log.Println("Unable to instantiate DB connection in /html/toggle_visibility: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
	}

	button := components.VisibilityIndicator(experiment_id, visibility)

	if err := button.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render visibility indicator", http.StatusInternalServerError)
		return
	}
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
