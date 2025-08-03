package handlers

import (
	"encoding/json"
	"net/http"

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

	id, err := database.CreatePackage(input.Name, userID)
	if err != nil {
		http.Error(w, "package creation failed", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": id, "name": input.Name})
}
