package handlers

import (
	"encoding/json"
	"net/http"

	"bundlr/internal/auth"
	"bundlr/internal/database"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(creds.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	id, err := database.CreateUser(creds.Email, hash)
	if err != nil {
		http.Error(w, "error creating user", http.StatusBadRequest)
		return
	}

	token, _ := auth.GenerateToken(id)

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	id, hash, err := database.GetUserByEmail(creds.Email)
	if err != nil || !auth.CheckPassword(hash, creds.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := auth.GenerateToken(id)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
