package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"bundlr/internal/auth"
	"bundlr/internal/config"
	"bundlr/internal/database"
	"bundlr/internal/handlers"
	"bundlr/internal/storage"
)

func main() {
	cfg := config.Load()

	storage.InitMinIO()

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
	r.Group(func(protected chi.Router) {
		protected.Use(auth.AuthMiddleware)
		protected.Post("/packages", handlers.CreatePackage)
		protected.Post("/packages/{name}/versions", handlers.CreateVersion)
	})
	r.Get("/packages/{name}/versions/{version}/download", handlers.DownloadVersion)

	fmt.Println("🚀 Bundlr running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
