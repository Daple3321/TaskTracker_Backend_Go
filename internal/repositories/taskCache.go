package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/redis/go-redis/v9"
)

type TaskCache struct {
	rdb *redis.Client
}

func NewTaskCache(rdb *redis.Client) *TaskCache {
	return &TaskCache{rdb: rdb}
}

func (t *TaskCache) GetTask(ctx context.Context, userID, taskID int) (*entity.Task, bool, error) {

	key := fmt.Sprintf("task:%d:%d", userID, taskID)

	pipe := t.rdb.Pipeline()
	getName := pipe.HGet(ctx, key, "name")
	getDescription := pipe.HGet(ctx, key, "description")
	getCreatedAt := pipe.HGet(ctx, key, "createdAt")
	getUpdatedAt := pipe.HGet(ctx, key, "updatedAt")

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, false, err
	}

	name, err := getName.Val(), getName.Err()
	if err != nil {
		return nil, false, err
	}

	description, err := getDescription.Val(), getDescription.Err()
	if err != nil {
		return nil, false, err
	}

	createdAt_Str, err := getCreatedAt.Val(), getCreatedAt.Err()
	if err != nil {
		return nil, false, err
	}
	createdAt, err := time.Parse(time.RFC3339, createdAt_Str)

	updatedAt_Str, err := getUpdatedAt.Val(), getUpdatedAt.Err()
	if err != nil {
		return nil, false, err
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAt_Str)

	finalTask := entity.Task{
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		UserId:      userID,
		Id:          taskID,
	}

	return &finalTask, true, nil
}

func (t *TaskCache) SetTask(ctx context.Context, task *entity.Task, ttl time.Duration) error {
	key := fmt.Sprintf("task:%d:%d", task.UserId, task.Id)

	pipe := t.rdb.Pipeline()
	setName := pipe.Set(ctx, key, task.Name, ttl)
	setDescription := pipe.Set(ctx, key, task.Description, ttl)
	setCreatedAt := pipe.Set(ctx, key, task.CreatedAt, ttl)
	setUpdatedAt := pipe.Set(ctx, key, task.UpdatedAt, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	_, err = setName.Result()
	if err != nil {
		return err
	}

	_, err = setDescription.Result()
	if err != nil {
		return err
	}

	_, err = setCreatedAt.Result()
	if err != nil {
		return err
	}

	_, err = setUpdatedAt.Result()
	if err != nil {
		return err
	}

	return nil
}

func (t *TaskCache) DeleteTask(ctx context.Context, userID, taskID int) error {
	_, err := t.rdb.Del(ctx, fmt.Sprintf("task:%d:%d", userID, taskID)).Result()
	if err != nil {
		return err
	}

	return nil
}

// func (t *TaskCache) DeleteTaskLists(ctx context.Context, userID int) error {
// }
