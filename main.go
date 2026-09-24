package main

import (
	filehandlers "XAI-liacs/BLADE_db/file_handlers"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"uuid"

	_ "github.com/lib/pq"
)

const maxUploadSize = 16 << 30 // 2 GB.

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

	jsondata, err := json.Marshal(health)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Println(r.RemoteAddr, r.Pattern, r.Body)
	w.Write(jsondata)
}
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

func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Protect against accidentally huge uploads.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	defer r.Body.Close()
	// Create a filename like:
	// blade_db_550e8400.zip
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
	// _ = os.Remove(path)
	unzipped_path := filepath.Base(path)

	unzip(path, unzipped_path)
	errs := filehandlers.ImportAllExperimentUnder(unzipped_path, "deployment")
	if len(errs) != 0 {
		http.Error(w, "failed to save to db.", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Received upload: %s\n", path)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "uploaded: %s\n", path)

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/submit", submitHandler)

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
