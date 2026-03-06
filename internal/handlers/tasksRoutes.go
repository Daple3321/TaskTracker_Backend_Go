package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/Daple3321/TaskTracker/internal/middleware"
	"github.com/Daple3321/TaskTracker/internal/repositories"
	"github.com/Daple3321/TaskTracker/internal/services"
	"github.com/Daple3321/TaskTracker/utils"
)

const requestTimeout = 3 * time.Second

var taskService *services.TaskService

type TasksHandler struct {
}

func NewHandler(db *sql.DB) *TasksHandler {

	taskRepo := repositories.NewTaskRepository(db)
	taskService = services.NewTaskService(taskRepo)

	return &TasksHandler{}
}

func (h *TasksHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /", middleware.Logging(middleware.Auth(GetTasksPaginated)))
	r.HandleFunc("GET /{id}/", middleware.Logging(GetTask))
	r.HandleFunc("POST /", middleware.Logging(CreateTask))
	r.HandleFunc("PUT /{id}/", middleware.Logging(UpdateTask))
	r.HandleFunc("DELETE /{id}/", middleware.Logging(DeleteTask))
	r.HandleFunc("GET /test/", middleware.Logging(TestRoute))

	return r
}

func TestRoute(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	err := taskService.TestFunc(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetTasksPaginated(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	response, err := taskService.GetTasksPaginated(ctx, pageStr, limitStr)
	if err != nil {
		if errors.Is(err, services.ErrNoPageParameter) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	fetchedTask, err := taskService.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidTask) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, repositories.ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.WriteJSONResponse(w, http.StatusOK, fetchedTask)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var newTask entity.Task
	err := utils.ParseJSON(r, &newTask)
	if err != nil {
		slog.Warn("[CreateTask] Error parsing request body", "err", err)
		//utils.WriteJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Error parsing request body. %s", err))
		http.Error(w, fmt.Sprintf("Error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	newTask.CreatedAt = time.Now()
	newTask.UpdatedAt = time.Now()

	createdId, err := taskService.AddTask(ctx, &newTask)
	if err != nil {
		if errors.Is(err, services.ErrInvalidTask) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	newTask.Id = createdId

	slog.Info("[CreateTask] New task created", "taskId", createdId)
	utils.WriteJSONResponse(w, http.StatusCreated, newTask)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	idString := r.PathValue("id")
	id, _ := strconv.Atoi(idString)

	var requestTask entity.Task
	err := utils.ParseJSON(r, &requestTask)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	updatedTask, err := taskService.UpdateTask(ctx, id, &requestTask)
	if err != nil {
		slog.Error("[UpdateTask] Error updating task", "err", err.Error())
		if errors.Is(err, repositories.ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if errors.Is(err, services.ErrInvalidTask) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	slog.Info("[UpdateTask] Task updated", "taskId", id)
	utils.WriteJSONResponse(w, http.StatusOK, updatedTask)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	err := taskService.DeleteTask(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if errors.Is(err, services.ErrInvalidTask) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	slog.Info("[DeleteTask] Task succesfuly deleted", "taskId", id)
	utils.WriteJSONResponse(w, http.StatusOK, fmt.Sprintf("Task {%d} succesfuly deleted", id))
}
