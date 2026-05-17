package main

import (
	"log"
	"net/http"

	"github.com/sssurendra99/mini-saas-backend/internal/config"
	"github.com/sssurendra99/mini-saas-backend/internal/db"
	"github.com/sssurendra99/mini-saas-backend/internal/tasks"
)

func main() {

	config.Load()

	dbPool := db.NewDBConnection()

	mux := http.NewServeMux()

	taskHandlers := tasks.BuildTaskModule(dbPool)

	tasks.RegisterRoutes(mux, taskHandlers)

	server := &http.Server{
		Addr:    ":9000",
		Handler: mux,
	}

	log.Println("Server running on http://localhost:9000")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
