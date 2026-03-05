package services

import (
	"errors"
	"strings"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/Daple3321/TaskTracker/internal/repositories"
)

var ErrInvalidTask = errors.New("invalid task")

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

func (t *TaskService) GetTasks() ([]entity.Task, error) {
	return t.storage.GetAllTasks()
}

func (t *TaskService) GetTask(taskId int) (*entity.Task, error) {

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
