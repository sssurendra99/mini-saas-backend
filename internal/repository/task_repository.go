package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sssurendra99/mini-saas-backend/internal/domain"
)

type TaskRepository struct {
	dbConnection *pgxpool.Pool
}

func NewTaskRepository(dbConnection *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		dbConnection: dbConnection,
	}
}

func (r *TaskRepository) GetAll() ([]domain.Task, error) {
	var tasks []domain.Task

	rows, err := r.dbConnection.Query(
		context.Background(),
		"SELECT id, title, completed, user_id FROM tasks;",
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task domain.Task

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Completed,
			&task.UserId,
		)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}