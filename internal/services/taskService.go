package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/Daple3321/TaskTracker/internal/middleware"
	"github.com/Daple3321/TaskTracker/internal/repositories"
)

var ErrInvalidTask = errors.New("invalid task")
var ErrNoPageParameter = errors.New("no page parameter specified")

type TaskService struct {
	storage     repositories.Repository
	TagsStorage *repositories.TagsRepository
}

func NewTaskService(tasksRepo repositories.Repository, tagsRepo *repositories.TagsRepository) *TaskService {

	ts := TaskService{
		storage:     tasksRepo,
		TagsStorage: tagsRepo,
	}

	return &ts
}

func (t *TaskService) TestFunc(ctx context.Context) error {
	select {
	case <-time.After(3 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func GetUserIdFromCtx(ctx context.Context) (int, error) {
	userIdVal := ctx.Value(middleware.ContextUserIdKey)
	if userIdVal == nil {
		return 0, errors.New("user id not in context")
	}
	userId, ok := userIdVal.(int)
	if !ok {
		return 0, errors.New("user id invalid type")
	}

	return userId, nil
}

func (t *TaskService) GetTasksCount(ctx context.Context) (int, error) {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	cnt, err := t.storage.GetTasksCount(ctx, userId)

	return cnt, err
}

func (t *TaskService) GetAllTasks(ctx context.Context) ([]entity.Task, error) {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	allTasks, err := t.storage.GetAllTasks(ctx, userId)
	if err != nil {
		return nil, err
	}

	result, err := t.PopulateTagsForTasks(ctx, allTasks)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TaskService) GetTasksPaginated(ctx context.Context, pageStr string, limitStr string) (*entity.PaginatedResponse, error) {

	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return nil, err
	}

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

	tasks, err := t.storage.GetTasksPaginated(ctx, userId, page, limit)
	if err != nil {
		return nil, err
	}

	result, err := t.PopulateTagsForTasks(ctx, tasks)
	if err != nil {
		return nil, err
	}

	totalItems, err := t.GetTasksCount(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := (totalItems + limit - 1) / limit

	response := entity.PaginatedResponse{
		Items:      result,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return &response, nil
}

func (t *TaskService) GetTask(ctx context.Context, taskId int) (*entity.Task, error) {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if taskId < 0 {
		return nil, ErrInvalidTask
	}

	fetchedTask, err := t.storage.GetTask(ctx, userId, taskId)
	if err != nil {
		return nil, err
	}

	// fetch tags for task
	tags, err := t.TagsStorage.GetTagsForTask(ctx, userId, taskId)
	if err != nil {
		return nil, err
	}
	fetchedTask.Tags = tags

	return fetchedTask, nil
}

func (t *TaskService) AddTask(ctx context.Context, newTask *entity.Task) (int, error) {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	newTask.Name = strings.TrimSpace(newTask.Name)
	newTask.Description = strings.TrimSpace(newTask.Description)
	if newTask.Name == "" {
		return 0, ErrInvalidTask
	}

	return t.storage.CreateTask(ctx, userId, newTask)
}

func (t *TaskService) UpdateTask(ctx context.Context, id int, updatedTask *entity.Task) (*entity.Task, error) {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	updatedTask.Name = strings.TrimSpace(updatedTask.Name)
	updatedTask.Description = strings.TrimSpace(updatedTask.Description)
	if updatedTask.Name == "" {
		return nil, ErrInvalidTask
	}

	fetchedTask, err := t.storage.UpdateTask(ctx, userId, id, updatedTask)
	if err != nil {
		return nil, err
	}

	result, err := t.PopulateTagsForTasks(ctx, []entity.Task{*fetchedTask})
	if err != nil {
		return nil, err
	}

	return &result[0], nil
}

func (t *TaskService) DeleteTask(ctx context.Context, id int) error {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return err
	}

	return t.storage.DeleteTask(ctx, userId, id)
}

// TODO: Bad. Making and copying redundant slices
func (t *TaskService) PopulateTagsForTasks(ctx context.Context, tasks []entity.Task) ([]entity.Task, error) {
	userId, err := GetUserIdFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	result := []entity.Task{}

	for _, task := range tasks {
		tags, err := t.TagsStorage.GetTagsForTask(ctx, userId, task.Id)
		if err != nil {
			return nil, err
		}

		task.Tags = tags

		result = append(result, task)
	}

	return result, nil
}
