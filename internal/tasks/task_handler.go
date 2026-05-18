package tasks

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sssurendra99/mini-saas-backend/internal/response"
	"github.com/sssurendra99/mini-saas-backend/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetAllTasks()

	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		log.Fatalln(err)
		return
	}
	response.WriteJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {

	var uuid pgtype.UUID

	id := r.PathValue("id")
	err := uuid.Scan(id)

	if err != nil {
		http.Error(w, "Invalid task Id!", http.StatusBadRequest)
		log.Println(err)
		return
	}

	task, err := h.service.GetTaskById(uuid)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			log.Println(err)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		log.Println(err)
		return
	}

	response.WriteJSON(w, http.StatusOK, task)

}
