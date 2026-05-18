package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sssurendra99/mini-saas-backend/internal/domain"
	"github.com/sssurendra99/mini-saas-backend/internal/repository"
)

type TaskService struct {
	repo *repository.TaskRepository
}

func NewTaskService(tr *repository.TaskRepository) *TaskService {
	return &TaskService{
		repo: tr,
	}
}

func (s *TaskService) GetAllTasks() ([]domain.Task, error) {
	return s.repo.GetAll()
}

func (s *TaskService) GetTaskById(id pgtype.UUID) (domain.Task, error) {
	return s.repo.GetTaskById(id)
}
