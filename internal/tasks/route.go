package tasks

import (
	"net/http"
)

func RegisterRoutes(
	mux *http.ServeMux,
	taskHandler *TaskHandler,
){
	mux.HandleFunc("/tasks", taskHandler.GetAllTasks)
}