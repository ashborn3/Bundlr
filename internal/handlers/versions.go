package handlers

import (
	"encoding/json"
	"net/http"

	"bundlr/internal/auth"
	"bundlr/internal/database"
	"bundlr/internal/storage"

	"github.com/go-chi/chi/v5"
)

type VersionInput struct {
	Version string `json:"version"`
}

func CreateVersion(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	packageName := chi.URLParam(r, "name")

	pkgID, ownerID, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "not authorized", http.StatusForbidden)
		return
	}

	var input VersionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Version == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	fileKey := "uploads/" + packageName + "/" + input.Version + ".tar.gz"
	uploadURL, err := storage.GeneratePresignedUpload(fileKey)
	if err != nil {
		http.Error(w, "failed to generate upload url", http.StatusInternalServerError)
		return
	}

	id, err := database.CreateVersion(pkgID, input.Version, fileKey)
	if err != nil {
		http.Error(w, "failed to create version", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":         id,
		"version":    input.Version,
		"file_key":   fileKey,
		"upload_url": uploadURL,
	})
}
