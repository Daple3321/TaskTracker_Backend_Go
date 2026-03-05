package handlers

import (
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

	r.HandleFunc("GET /", middleware.Logging(middleware.Auth(GetTasks)))
	r.HandleFunc("GET /{id}/", middleware.Logging(GetTask))
	r.HandleFunc("POST /", middleware.Logging(CreateTask))
	r.HandleFunc("PUT /{id}/", middleware.Logging(UpdateTask))
	r.HandleFunc("DELETE /{id}/", middleware.Logging(DeleteTask))

	return r
}

// TODO: transfer pagination logic and handling to repo layer or service
func GetTasks(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Default to page 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10 // Default to 10 items per page
	}

	//totalItems := len(tasks)
	totalItems, err := taskService.GetTasksCount()
	allTasks, err := taskService.GetTasks()
	totalPages := (totalItems + limit - 1) / limit

	// Calculate offset and end index for the current page
	offset := (page - 1) * limit
	endIndex := offset + limit
	if endIndex > totalItems {
		endIndex = totalItems
	}

	// Get items for the current page
	var currentPageItems []entity.Task
	if offset < totalItems {
		currentPageItems = allTasks[offset:endIndex]
	}

	response := entity.PaginatedResponse{
		Items:      currentPageItems,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	// response := map[string]any{
	// 	"message": fmt.Sprintf("page={%d}, limit={%d}", page, limit),
	// 	"Tasks":   tasks,
	// }

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	fetchedTask, err := taskService.GetTask(id)
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

	createdId, err := taskService.AddTask(&newTask)
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
	idString := r.PathValue("id")
	id, _ := strconv.Atoi(idString)

	var requestTask entity.Task
	err := utils.ParseJSON(r, &requestTask)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	updatedTask, err := taskService.UpdateTask(id, &requestTask)
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
	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	err := taskService.DeleteTask(id)
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
