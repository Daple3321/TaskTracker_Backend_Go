package handlers

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Daple3321/ServerLearn/internal/entity"
	"github.com/Daple3321/ServerLearn/internal/middleware"
	"github.com/Daple3321/ServerLearn/internal/services"
	"github.com/Daple3321/ServerLearn/utils"
)

var taskService *services.TaskService

type TasksHandler struct {
}

func NewHandler() *TasksHandler {
	taskService = services.NewTaskService()
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

func GetTasks(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(2 * time.Second)

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
	totalItems := taskService.GetTasksCount()
	allTasks := taskService.GetTasks()
	totalPages := (totalItems + limit - 1) / limit // Calculate total pages

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

	// Construct the paginated response
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
	if id < 0 {
		slog.Info("[GetTask] Task not found", "taskId", id)
		//utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found", id))
		http.Error(w, fmt.Sprintf("Task with ID {%d} not found", id), http.StatusNotFound)
		return
	}

	fetchedTask, err := taskService.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// this madness is because relative path from cmd/api/main.go to /templates/
	// is ../../templates/task.html
	tmpl, err := template.ParseFiles(filepath.Join("..", "..", "templates", "task.html"))
	if err != nil {
		log.Printf("[GetTask] Error parsing html template: %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, fetchedTask)
	//utils.WriteJSONResponse(w, http.StatusOK, tasks[id])
}

func CreateTask(w http.ResponseWriter, r *http.Request) {

	//log.Printf("Remote addr: %s", r.RemoteAddr)

	var newTask entity.Task
	err := utils.ParseJSON(r, &newTask)
	if err != nil {
		log.Printf("[CreateTask] Error parsing request body. %s", err)
		//utils.WriteJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Error parsing request body. %s", err))
		http.Error(w, fmt.Sprintf("Error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if newTask.Name == "" {
		log.Printf("[CreateTask] No task name specified.")
		//utils.WriteJSONResponse(w, http.StatusBadRequest, "No task name specified")
		http.Error(w, "No task name specified", http.StatusBadRequest)
		return
	}

	newTask.CreatedAt = time.Now()
	newTask.UpdatedAt = time.Now()
	//tasksMu.Lock()
	//defer tasksMu.Unlock()

	createdId := taskService.AddTask(&newTask)

	newTask.Id = createdId

	log.Printf("[CreateTask] New task created with ID: {%d}", createdId)
	utils.WriteJSONResponse(w, http.StatusCreated, newTask)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, _ := strconv.Atoi(idString)

	var requestTask entity.Task
	err := utils.ParseJSON(r, &requestTask)
	if err != nil {
		log.Printf("[UpdateTask] Error parsing request body. %s", err)
		//utils.WriteJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Error parsing request body. %s", err))
		http.Error(w, fmt.Sprintf("Error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if requestTask.Name == "" {
		log.Printf("[UpdateTask] No task name specified.")
		//utils.WriteJSONResponse(w, http.StatusBadRequest, "No task name specified")
		http.Error(w, "No task name specified", http.StatusBadRequest)
		return
	}

	//tasksMu.Lock()
	//defer tasksMu.Unlock()

	updatedTask, err := taskService.UpdateTask(id, &requestTask)
	if err != nil {
		log.Printf("[UpdateTask] %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[UpdateTask] Task {%d} updated.", id)
	utils.WriteJSONResponse(w, http.StatusOK, updatedTask)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	//tasksMu.Lock()
	//defer tasksMu.Unlock()

	err := taskService.DeleteTask(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[DeleteTask] Task {%d} succesfuly deleted", id)
	utils.WriteJSONResponse(w, http.StatusOK, fmt.Sprintf("Task {%d} succesfuly deleted", id))
}
