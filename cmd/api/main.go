package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/sssurendra99/mini-saas-backend/internal/config"
	"github.com/sssurendra99/mini-saas-backend/internal/db"
	"github.com/sssurendra99/mini-saas-backend/internal/middleware"
	"github.com/sssurendra99/mini-saas-backend/internal/tasks"
)

func main() {

	config.Load()

	dbPool := db.NewDBConnection()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	taskHandlers := tasks.BuildTaskModule(dbPool)

	tasks.RegisterRoutes(mux, taskHandlers)

	loggedMux := middleware.LoggingBehavior(mux) // This will return a handler because the LoggingBehaviour also returns a handler

	server := &http.Server{
		Addr:    ":9000",
		Handler: loggedMux,
	}

	log.Println("Server running on http://localhost:9000")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
