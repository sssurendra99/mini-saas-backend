package tasks

import (
	"log"
	"net/http"

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