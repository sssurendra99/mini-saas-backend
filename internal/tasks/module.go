package tasks

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sssurendra99/mini-saas-backend/internal/repository"
	"github.com/sssurendra99/mini-saas-backend/internal/service"
)

func BuildTaskModule(db *pgxpool.Pool) *TaskHandler {
	repo := repository.NewTaskRepository(db)
	service := service.NewTaskService(repo)
	handler := NewTaskHandler(service)
	return handler
}
