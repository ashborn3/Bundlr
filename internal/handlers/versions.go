package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"bundlr/internal/database"
	"bundlr/internal/storage"
	"bundlr/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

type VersionInput struct {
	Version  string `json:"version"`
	FileName string `json:"file_name"`
}

func DownloadVersion(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")

	pkgID, _, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}

	fileVersion, err := database.GetVersion(pkgID, version)
	if err != nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	downloadURL, err := storage.GeneratePresignedDownload(fileVersion.FileKey)
	if err != nil {
		http.Error(w, "failed to generate download URL", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"download_url": downloadURL,
		"file_name":    fileVersion.FileName,
	})

}

func GetUploadURL(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")

	var input struct {
		Version  string `json:"version"`
		FileName string `json:"file_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	if input.Version == "" || input.FileName == "" {
		http.Error(w, "version and file_name are required", http.StatusBadRequest)
		return
	}

	// Generate safe file key
	fileKey := utils.MakeFileKey(packageName, input.Version, input.FileName)

	// Generate presigned URL
	uploadURL, err := storage.GeneratePresignedUpload(fileKey)
	if err != nil {
		http.Error(w, "failed to generate upload URL", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"file_key":   fileKey,
		"upload_url": uploadURL,
	})
}

func ConfirmVersionUpload(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")

	var input struct {
		Version  string `json:"version"`
		FileKey  string `json:"file_key"`
		FileName string `json:"file_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	if input.Version == "" || input.FileKey == "" || input.FileName == "" {
		http.Error(w, "version, file_key and file_name are required", http.StatusBadRequest)
		return
	}

	_, err := storage.Client.StatObject(
		context.Background(),
		storage.Bucket,
		input.FileKey,
		minio.StatObjectOptions{},
	)
	if err != nil {
		http.Error(w, "file not uploaded", http.StatusBadRequest)
		return
	}

	// ✅ Get package ID
	pkgID, _, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}

	// ✅ Insert version into DB
	versionID, err := database.CreateVersion(pkgID, input.Version, input.FileKey, input.FileName)
	if err != nil {
		http.Error(w, "failed to create version", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":        versionID,
		"version":   input.Version,
		"file_name": input.FileName,
	})
}
