package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/Daple3321/TaskTracker/internal/repositories"
)

var ErrInvalidTask = errors.New("invalid task")
var ErrNoPageParameter = errors.New("no page parameter specified")

type TaskService struct {
	storage repositories.Repository
}

func NewTaskService(storage repositories.Repository) *TaskService {

	ts := TaskService{
		storage: storage,
	}

	return &ts
}

func (t *TaskService) GetTasksCount() (int, error) {

	cnt, err := t.storage.GetTasksCount()

	return cnt, err
}

func (t *TaskService) GetAllTasks() ([]entity.Task, error) {
	return t.storage.GetAllTasks()
}

func (t *TaskService) GetTasksPaginated(pageStr string, limitStr string) (*entity.PaginatedResponse, error) {

	pageStr = strings.TrimSpace(pageStr)
	if pageStr == "" {
		return nil, ErrNoPageParameter
	}
	limitStr = strings.TrimSpace(limitStr)

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Default to page 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10 // Default to 10 items per page
	}

	tasks, err := t.storage.GetTasksPaginated(page, limit)
	if err != nil {
		return nil, err
	}

	totalItems, err := t.GetTasksCount()
	if err != nil {
		return nil, err
	}

	totalPages := (totalItems + limit - 1) / limit

	response := entity.PaginatedResponse{
		Items:      tasks,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return &response, nil
}

func (t *TaskService) GetTask(taskId int) (*entity.Task, error) {

	if taskId < 0 {
		return nil, ErrInvalidTask
	}

	fetchedTask, err := t.storage.GetTask(taskId)
	if err != nil {
		return nil, err
	}

	return fetchedTask, nil
}

func (t *TaskService) AddTask(newTask *entity.Task) (int, error) {

	newTask.Name = strings.TrimSpace(newTask.Name)
	newTask.Description = strings.TrimSpace(newTask.Description)
	if newTask.Name == "" {
		return 0, ErrInvalidTask
	}

	return t.storage.CreateTask(newTask)
}

func (t *TaskService) UpdateTask(id int, updatedTask *entity.Task) (*entity.Task, error) {

	updatedTask.Name = strings.TrimSpace(updatedTask.Name)
	updatedTask.Description = strings.TrimSpace(updatedTask.Description)
	if updatedTask.Name == "" {
		return nil, ErrInvalidTask
	}

	return t.storage.UpdateTask(id, updatedTask)
}

func (t *TaskService) DeleteTask(id int) error {
	return t.storage.DeleteTask(id)
}
