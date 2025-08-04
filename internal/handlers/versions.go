package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"bundlr/internal/auth"
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

	userID := auth.GetUserID(r)

	_, ownerID, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
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

	stat, err := storage.Client.StatObject(context.Background(), storage.Bucket, input.FileKey, minio.StatObjectOptions{})
	if err != nil {
		http.Error(w, "file not uploaded", http.StatusBadRequest)
		return
	}

	pkgID, _, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}

	versionID, err := database.CreateVersion(pkgID, input.Version, input.FileKey, input.FileName, stat.Size)
	if err != nil {
		http.Error(w, "failed to create version", http.StatusInternalServerError)
		return
	}

	database.IncrementDownloadCount(pkgID, versionID)

	json.NewEncoder(w).Encode(map[string]string{
		"id":        versionID,
		"version":   input.Version,
		"file_name": input.FileName,
	})
}

func ListVersions(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")

	pkgID, _, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}

	versions, err := database.ListVersions(pkgID)
	if err != nil {
		http.Error(w, "failed to fetch versions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(versions)
}

func DeleteVersion(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	packageName := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")

	pkgID, ownerID, err := database.GetPackageByName(packageName)
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Optional: Delete from MinIO
	v, err := database.GetVersion(pkgID, version)
	if err == nil {
		storage.Client.RemoveObject(context.Background(), storage.Bucket, v.FileKey, minio.RemoveObjectOptions{})
	}

	if err := database.DeleteVersion(pkgID, version); err != nil {
		http.Error(w, "failed to delete version", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
