package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bundlr/internal/auth"
	"bundlr/internal/database"
)

type PackageInput struct {
	Name string `json:"name"`
}

func CreatePackage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var input PackageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	input.Name = strings.ToLower(input.Name)

	id, err := database.CreatePackage(input.Name, userID)
	if err != nil {
		http.Error(w, "package creation failed", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": id, "name": input.Name})
}

func ListPackages(w http.ResponseWriter, r *http.Request) {
	pkgs, err := database.ListPackages()
	if err != nil {
		http.Error(w, "failed to fetch packages", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(pkgs)
}

func SearchPackagesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 10
	}

	pkgs, err := database.SearchPackages(query, limit, offset)
	if err != nil {
		http.Error(w, "failed to fetch packages", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(pkgs)
}
