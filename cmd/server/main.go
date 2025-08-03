package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"bundlr/internal/config"
	"bundlr/internal/database"
	"bundlr/internal/handlers"
)

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Post("/auth/register", handlers.Register)
	r.Post("/auth/login", handlers.Login)

	fmt.Println("🚀 Bundlr running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
