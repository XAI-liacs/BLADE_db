package main

import (
	"XAI-liacs/BLADE_db/authentication"
	"XAI-liacs/BLADE_db/components"
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

// Pass with write context...
func (app *App) signUpHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	transactional_state, err := app.State.GetTransactionalState(r.Context())

	if err != nil {
		http.Error(w, "Unable to start transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	_, err = app.State.DB.GetUserNamed(r.Context(), username)
	if err == nil {
		http.Error(w, "User already exists.", http.StatusBadRequest)
		return
	}
	pass_hash := sha256.Sum256([]byte(password))

	api_key, err := GenerateAPIKey()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = transactional_state.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		Name:         username,
		PasswordHash: pass_hash[:],
		ApiKey:       api_key,
		IsAdmin:      false,
	})
	if err != nil {
		http.Error(w, "Unable to create user, "+err.Error(), http.StatusInternalServerError)
		transactional_state.TX.Rollback()
		return
	}
	if err := transactional_state.TX.Commit(); err != nil {
		http.Error(w, "Unable to commit user creation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	db_user, err := app.State.DB.GetUserNamed(r.Context(), username)
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := app.authenticate(r, "html")
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

func (app *App) HandleHtmlExperiment(w http.ResponseWriter, r *http.Request) {
	component := components.EmptyExperiment()
	user, err := app.authenticate(r, "html")
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	experiments, err := app.State.DB.GetExperimentsForUser(context.Background(), user.ID)
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

func (app *App) HandleRunCells(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticate(r, "html")
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

	ownership_map, err := app.State.DB.GetUsersAndVisibilityOfExperiment(r.Context(), experiment_id)
	if err != nil {
		log.Println("Got error loading viewability in /html/run: ", err.Error())
		http.Error(w, "DB error.", http.StatusInternalServerError)
		return
	}
	if !(slices.Contains(ownership_map.UserIds, user.ID) || ownership_map.IsPublic) {
		http.Error(w, "Unauthorised access", http.StatusInternalServerError)
		return
	}

	runs, err := app.State.DB.GetRunsForExperiment(r.Context(), experiment_id)
	if err != nil {
		log.Println("Got error loading runs for expreriment "+experiment_id.String()+" in /html/run: ", err.Error())
		http.Error(w, "DB access error.", http.StatusInternalServerError)
		return
	}
	runCells := components.RunCells(runs, experiment_id, slices.Contains(ownership_map.UserIds, user.ID), ownership_map.IsPublic)
	runCells.Render(r.Context(), w)
}

func (app *App) HandleHtmlDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticate(r, "html")
	if err != nil {
		http.Error(w, "Unable to authorise user: "+err.Error(), http.StatusUnauthorized)
		return
	}
	user_dashboard := components.Dashboard(user)
	user_dashboard.Render(r.Context(), w)
}

func (app *App) HandleHtmlDefaultUserVisibility(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticate(r, "html")
	if err != nil {
		http.Error(w, "Unable to authorise user: "+err.Error(), http.StatusUnauthorized)
		return
	}
	writerContext, err := app.State.GetTransactionalState(r.Context())
	defer writerContext.TX.Commit()
	if err != nil {
		http.Error(w, "Unable to get DB writer context: "+err.Error(), http.StatusInternalServerError)
		writerContext.TX.Rollback()
		return
	}

	if err = writerContext.DB.SetDefaultVisibility(r.Context(), database.SetDefaultVisibilityParams{
		ID:                user.ID,
		DefaultVisibility: !user.DefaultVisibility,
	}); err != nil {
		http.Error(w, "Unable to get update default visibility: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user_dashboard := components.DefaultVisibility(!user.DefaultVisibility) //Relect DB change in fetched object..
	user_dashboard.Render(r.Context(), w)
}

func (app *App) handleChangeAPIKey(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticate(r, "html")
	if err != nil {
		http.Error(w, "Unauthorised access: "+err.Error(), http.StatusUnauthorized)
		return
	}
	api, err := GenerateAPIKey()
	if err != nil {
		http.Error(w, "Unable to generate key: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writerContext, err := app.State.GetTransactionalState(r.Context())
	defer writerContext.TX.Commit()

	if err != nil {
		writerContext.TX.Rollback()
		http.Error(w, "Unable to generate key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err = writerContext.DB.UpdateAPIKey(r.Context(), database.UpdateAPIKeyParams{
		ID:     user.ID,
		ApiKey: api,
	}); err != nil {
		writerContext.TX.Rollback()
		http.Error(w, "Unable to set api key: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(api))
}

func (app *App) HandleHtmlRunData(w http.ResponseWriter, r *http.Request) {
	component := components.EmptyRightPanel("No runs selected.")

	app.authenticate(r, "html")

	runID, err := uuid.Parse(r.URL.Query().Get("run_id"))
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	solutions, err := app.State.DB.GetSolutionsForRun(r.Context(), runID)
	if err != nil {
		component.Render(r.Context(), w)
		return
	}

	conversationLog, err := app.State.DB.GetConversationLog(r.Context(), runID)
	if err != nil {
		conversationLog = []database.GetConversationLogRow{}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	components.RunData(runID, solutions, conversationLog).Render(r.Context(), w)
}

func (app *App) HandleHtmlVisibilityToggle(w http.ResponseWriter, r *http.Request) {
	_, err := app.authenticate(r, "html")
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

	transaction_state, err := app.State.GetTransactionalState(r.Context())
	if err != nil {
		log.Println("Unable to instantiate transaction DB connection in /html/toggle_visibility: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer transaction_state.TX.Commit()

	visibility, err := app.State.DB.ToggleExperimentVisibility(r.Context(), experiment_id)
	if err != nil {
		app.State.TX.Rollback()
		log.Println("Unable to instantiate DB connection in /html/toggle_visibility: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
	}

	button := components.VisibilityIndicator(experiment_id, visibility)

	if err := button.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render visibility indicator", http.StatusInternalServerError)
		return
	}
}

func (app *App) HandleHtmlContributors(w http.ResponseWriter, r *http.Request) {
	_, err := app.authenticate(r, "html")
	if err != nil {
		log.Println("Got unauthorised access to /html/contributors" + err.Error())
		http.Error(w, "Unauthorised access.", http.StatusUnauthorized)
		return
	}
	run_id, err := uuid.Parse(r.URL.Query().Get("run_id"))
	if err != nil {
		log.Println("Query error in /html/contributors" + err.Error())
		http.Error(w, "Bad access error, expected /html/contributors?run_id=<some uuid>", http.StatusBadRequest)
		return
	}

	methods, err := app.State.DB.GetMethodsForRun(r.Context(), run_id)
	if err != nil {
		log.Println("Unable to get methods for /html/contributors: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	llms, err := app.State.DB.GetLLMForRun(r.Context(), run_id)
	if err != nil {
		log.Println("Unable to get LLMs for /html/contributors: " + err.Error())
		http.Error(w, "Internal DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	modalView := components.ContributorsView(methods, llms)
	if err := modalView.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render visibility indicator", http.StatusInternalServerError)
		return
	}
}
