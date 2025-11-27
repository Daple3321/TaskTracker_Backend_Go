package tasks

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"gameroll.com/ServerLearn/utils"
)

var tasks []Task = []Task{
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

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /", GetTasks)
	r.HandleFunc("GET /{id}/", GetTask)
	r.HandleFunc("POST /", CreateTask)
	r.HandleFunc("PUT /{id}/", UpdateTask)
	r.HandleFunc("DELETE /{id}/", DeleteTask)

	return r
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("GetTasks/")
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
	var currentPageItems []Task
	if offset < totalItems {
		currentPageItems = tasks[offset:endIndex]
	}

	// Construct the paginated response
	response := PaginatedResponse{
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
	log.SetPrefix("GetTask/")

	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)
	if id >= len(tasks) {
		log.Printf("Task with ID {%d} not found", id)
		utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found", id))
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, tasks[id])
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("CreateTask/")

	var newTask Task
	utils.ParseJSON(r, &newTask)
	defer r.Body.Close()

	if newTask.Name == "" {
		log.Printf("No task name specified.")
		utils.WriteJSONResponse(w, http.StatusBadRequest, "No task name specified")
		return
	}

	newTask.Id = len(tasks) + 1
	newTask.CreatedAt = time.Now()
	newTask.UpdatedAt = time.Now()

	tasks = append(tasks, newTask)

	log.Printf("New task created with ID: {%d}", newTask.Id)
	utils.WriteJSONResponse(w, http.StatusCreated, tasks[len(tasks)-1])
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("UpdateTask/")

	idString := r.PathValue("id")
	id, _ := strconv.Atoi(idString)
	idx := slices.IndexFunc(tasks, func(t Task) bool {
		return t.Id == id
	})
	if idx == -1 {
		log.Printf("Task with ID {%d} not found.", id)
		utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found.", id))
		return
	}

	var updatedTask Task
	utils.ParseJSON(r, &updatedTask)
	defer r.Body.Close()

	if updatedTask.Name == "" {
		log.Printf("No task name specified.")
		utils.WriteJSONResponse(w, http.StatusBadRequest, "No task name specified")
		return
	}

	tasks[idx].Name = updatedTask.Name
	tasks[idx].Description = updatedTask.Description
	tasks[idx].UpdatedAt = time.Now()

	log.Printf("Task {%d} updated.", idx)
	utils.WriteJSONResponse(w, http.StatusOK, tasks[idx])
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("DeleteTask/")

	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	idx := slices.IndexFunc(tasks, func(t Task) bool {
		return t.Id == id
	})

	if idx != -1 {
		tasks = slices.Delete(tasks, idx, idx+1)
	} else {
		log.Printf("Task with ID {%d} not found", id)
		utils.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("Task with ID {%d} not found", id))
		return
	}

	log.Printf("Task {%d} succesfuly deleted", id)
	utils.WriteJSONResponse(w, http.StatusOK, fmt.Sprintf("Task {%d} succesfuly deleted", id))
}
