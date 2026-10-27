package main

import (
	"XAI-liacs/BLADE_db/authentication"
	"XAI-liacs/BLADE_db/internal/database"
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Returns authentication the information about the user that was logged in.
//
// ## Params
//
// r: Request handler, makes no changes to request handler, checks for Cookies if root_endpoint = "html" and Bearer if root_endpoit="api"
//
// root_endpoint: based on the root of the end-point, can be either "api" or "html".
//
// ## Returns:
//
// `user: database.User`: Returns the user data from the database.
// `is_valid: bool`: Return is the authentication token and associate user exists in db.
func (app *App) authenticate(r *http.Request, root_endpoint string) (user database.User, err error) {
	var token string
	fmt.Println("\t Authentication for " + r.URL.Path)
	if root_endpoint == "html" {
		token, err = authentication.GetCookie(r.Header)
	} else {
		token, err = authentication.GetBearerToken(r.Header)
	}
	if err != nil {
		return user, err
	}
	id, err := authentication.ValidateJWT(token)
	if err != nil {
		return user, err
	}

	user, err = app.State.DB.GetUserWithID(r.Context(), id)
	if err != nil {
		return user, err
	}
	return user, nil
}

func preTestClear(dbConfig App) error {
	fmt.Println("\t=== Clearing all tables ===")
	ctx := context.Background()
	// Clear Relation Tables First.....
	if err := dbConfig.State.DB.ClearUserExperiments(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearRunDescriptor(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearConversationLog(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearMethodLLM(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearTagProblem(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearExperimentRun(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearParentChild(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearRunSolution(ctx); err != nil {
		return err
	}

	// Clear Base Tables Later.....

	if err := dbConfig.State.DB.ClearUsers(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearLogBefore(ctx, time.Now()); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearTag(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearRun(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearLLM(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearProblem(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearMessage(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearExperiment(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearMethod(ctx); err != nil {
		return err
	}
	if err := dbConfig.State.DB.ClearSolution(ctx); err != nil {
		return err
	}
	return nil
}
