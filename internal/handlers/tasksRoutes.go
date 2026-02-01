package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"gameroll.com/ServerLearn/internal/entity"
	"gameroll.com/ServerLearn/internal/services"
	"gameroll.com/ServerLearn/utils"
)

var tasksMu sync.RWMutex
var tasks []entity.Task = []entity.Task{
	{Name: "TEST TASK", Description: "123123123", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 0},
	{Name: "SECOND TASJ", Description: "ASDASWRWAR", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 1},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
	{Name: "WOW", Description: "AWAW^AW^A", CreatedAt: time.Now(), UpdatedAt: time.Now(), Id: 2},
}

// Всё неправильно тут сделано.
// Сам сервис должен хендлить все запросы
// Это получается как какой-то контроллер лишний
// ===Ask GPT===
var taskService *services.TaskService

type TasksHandler struct {
}

func NewHandler() *TasksHandler {
	taskService = services.NewTaskService()
	return &TasksHandler{}
}

func (h *TasksHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /", Logging(GetTasks))
	r.HandleFunc("GET /{id}/", Logging(GetTask))
	r.HandleFunc("POST /", Logging(CreateTask))
	r.HandleFunc("PUT /{id}/", Logging(UpdateTask))
	r.HandleFunc("DELETE /{id}/", Logging(DeleteTask))

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

	totalItems := len(tasks)
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
		currentPageItems = tasks[offset:endIndex]
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
	if id >= len(tasks) || id < 0 {
		log.Printf("[GetTask] Task with ID {%d} not found", id)
		//utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found", id))
		http.Error(w, fmt.Sprintf("Task with ID {%d} not found", id), http.StatusNotFound)
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
	tmpl.Execute(w, tasks[id])
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

	tasksMu.Lock()
	defer tasksMu.Unlock()

	newTask.Id = len(tasks) + 1
	newTask.CreatedAt = time.Now()
	newTask.UpdatedAt = time.Now()

	tasks = append(tasks, newTask)

	log.Printf("[CreateTask] New task created with ID: {%d}", newTask.Id)
	utils.WriteJSONResponse(w, http.StatusCreated, tasks[len(tasks)-1])
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, _ := strconv.Atoi(idString)
	idx := slices.IndexFunc(tasks, func(t entity.Task) bool {
		return t.Id == id
	})
	if idx == -1 {
		log.Printf("[UpdateTask] Task with ID {%d} not found.", id)
		//utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found.", id))
		http.Error(w, fmt.Sprintf("Task with ID {%d} not found", id), http.StatusNotFound)
		return
	}

	var updatedTask entity.Task
	err := utils.ParseJSON(r, &updatedTask)
	if err != nil {
		log.Printf("[UpdateTask] Error parsing request body. %s", err)
		//utils.WriteJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Error parsing request body. %s", err))
		http.Error(w, fmt.Sprintf("Error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if updatedTask.Name == "" {
		log.Printf("[UpdateTask] No task name specified.")
		//utils.WriteJSONResponse(w, http.StatusBadRequest, "No task name specified")
		http.Error(w, "No task name specified", http.StatusBadRequest)
		return
	}

	tasksMu.Lock()
	defer tasksMu.Unlock()

	tasks[idx].Name = updatedTask.Name
	tasks[idx].Description = updatedTask.Description
	tasks[idx].UpdatedAt = time.Now()

	log.Printf("[UpdateTask] Task {%d} updated.", idx)
	utils.WriteJSONResponse(w, http.StatusOK, tasks[idx])
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	idx := slices.IndexFunc(tasks, func(t entity.Task) bool {
		return t.Id == id
	})

	tasksMu.Lock()
	defer tasksMu.Unlock()

	if idx != -1 {
		tasks = slices.Delete(tasks, idx, idx+1)
	} else {
		log.Printf("[DeleteTask] Task with ID {%d} not found", id)
		//utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found", id))
		http.Error(w, fmt.Sprintf("Task with ID {%d} not found", id), http.StatusNotFound)
		return
	}

	log.Printf("[DeleteTask] Task {%d} succesfuly deleted", id)
	utils.WriteJSONResponse(w, http.StatusOK, fmt.Sprintf("Task {%d} succesfuly deleted", id))
}

func Logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s : %s] [%s]", r.Method, r.RequestURI, r.RemoteAddr)
		f(w, r)
	}
}
