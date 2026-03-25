package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/Daple3321/TaskTracker/internal/middleware"
	"github.com/Daple3321/TaskTracker/internal/repositories"
	"github.com/Daple3321/TaskTracker/internal/services"
	"github.com/Daple3321/TaskTracker/utils"
)

const requestTimeout = 3 * time.Second

type TasksHandler struct {
	TaskService *services.TaskService
}

func NewTaskHandler(db *sql.DB) *TasksHandler {

	taskRepo := repositories.NewTaskRepository(db)
	tagsRepo := repositories.NewTagsRepository(db)
	taskService := services.NewTaskService(taskRepo, tagsRepo)

	return &TasksHandler{TaskService: taskService}
}

func (h *TasksHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /", middleware.Logging(middleware.RateLimit(middleware.Auth(h.GetTasksPaginated))))
	r.HandleFunc("GET /{id}/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.GetTask))))
	r.HandleFunc("POST /", middleware.Logging(middleware.RateLimit(middleware.Auth(h.CreateTask))))
	r.HandleFunc("PUT /{id}/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.UpdateTask))))
	r.HandleFunc("DELETE /{id}/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.DeleteTask))))

	r.HandleFunc("GET /tags/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.GetAllTags))))
	r.HandleFunc("POST /tags/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.CreateTag))))
	r.HandleFunc("DELETE /tags/{tagId}/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.DeleteTag))))
	// Literal "task/" avoids ServeMux conflict with POST /tags/ (/{id}/tags/ would allow id == "tags").
	r.HandleFunc("POST /task/{id}/tags/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.AddTagToTask))))
	r.HandleFunc("DELETE /task/{id}/tags/{tagId}/", middleware.Logging(middleware.RateLimit(middleware.Auth(h.DeleteTagFromTask))))

	//r.HandleFunc("GET /test/", middleware.Logging(middleware.Auth(h.TestRoute)))

	return r
}

func (h *TasksHandler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	userId, err := services.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tags, err := h.TaskService.TagsStorage.GetTagsForUser(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, tags)
}

func (h *TasksHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	userId, err := services.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tag entity.Tag
	if err := utils.ParseJSON(r, &tag); err != nil {
		http.Error(w, fmt.Sprintf("error parsing tag. %s", err), http.StatusBadRequest)
		return
	}
	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		http.Error(w, "tag name is required", http.StatusBadRequest)
		return
	}

	existing, err := h.TaskService.TagsStorage.FindTagByName(ctx, userId, tag.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existing != nil {
		http.Error(w, repositories.ErrTagAlreadyExists.Error(), http.StatusConflict)
		return
	}

	tagId, err := h.TaskService.TagsStorage.CreateTag(ctx, userId, tag.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tag.Id = tagId
	tag.UserId = userId
	utils.WriteJSONResponse(w, http.StatusCreated, tag)
}

func (h *TasksHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	userId, err := services.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tagIdStr := r.PathValue("tagId")
	tagId, err := strconv.Atoi(tagIdStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing tag id. %s", err), http.StatusBadRequest)
		return
	}

	err = h.TaskService.TagsStorage.DeleteTag(ctx, userId, tagId)
	if err != nil {
		if errors.Is(err, repositories.ErrTaskNotFound) {
			http.Error(w, "tag not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TasksHandler) AddTagToTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	userId, err := services.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	taskIdString := r.PathValue("id")
	taskId, err := strconv.Atoi(taskIdString)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing task id. %s", err), http.StatusBadRequest)
		return
	}

	var tag entity.Tag
	err = utils.ParseJSON(r, &tag)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing tag. %s", err), http.StatusBadRequest)
		return
	}

	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		http.Error(w, "tag name is required", http.StatusBadRequest)
		return
	}

	existing, err := h.TaskService.TagsStorage.FindTagByName(ctx, userId, tag.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tagId int
	if existing != nil {
		tagId = existing.Id
	} else {
		tagId, err = h.TaskService.TagsStorage.CreateTag(ctx, userId, tag.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.TaskService.TagsStorage.SetTagForTask(ctx, userId, taskId, tagId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, fmt.Sprintf("Tag {%d} succesfuly added to task {%d}", tagId, taskId))
}

func (h *TasksHandler) DeleteTagFromTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	userId, err := services.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	taskIdString := r.PathValue("id")
	taskId, err := strconv.Atoi(taskIdString)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing task id. %s", err), http.StatusBadRequest)
		return
	}

	tagIdString := r.PathValue("tagId")
	tagId, err := strconv.Atoi(tagIdString)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing tag id. %s", err), http.StatusBadRequest)
		return
	}

	err = h.TaskService.TagsStorage.DeleteTagFromTask(ctx, userId, tagId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, fmt.Sprintf("Tag {%d} succesfuly deleted from task {%d}", tagId, taskId))
}

func (h *TasksHandler) TestRoute(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	err := h.TaskService.TestFunc(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TasksHandler) GetTasksPaginated(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	response, err := h.TaskService.GetTasksPaginated(ctx, pageStr, limitStr)
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

func (h *TasksHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing task id. %s", err), http.StatusBadRequest)
		return
	}

	fetchedTask, err := h.TaskService.GetTask(ctx, id)
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

func (h *TasksHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
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

	createdId, err := h.TaskService.AddTask(ctx, &newTask)
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
	// TODO: need to somehow also set the UserId field

	slog.Info("[CreateTask] New task created", "taskId", createdId)
	utils.WriteJSONResponse(w, http.StatusCreated, newTask)
}

func (h *TasksHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
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

	updatedTask, err := h.TaskService.UpdateTask(ctx, id, &requestTask)
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

func (h *TasksHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	id, _ := strconv.Atoi(idString)

	err := h.TaskService.DeleteTask(ctx, id)
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
