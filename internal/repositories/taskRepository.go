package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/Daple3321/TaskTracker/internal/entity"
)

type Repository interface {
	GetTasksCount(ctx context.Context) (int, error)
	GetAllTasks(ctx context.Context) ([]entity.Task, error)
	GetTasksPaginated(ctx context.Context, offset int, limit int) ([]entity.Task, error)
	GetTask(ctx context.Context, taskId int) (*entity.Task, error)
	CreateTask(ctx context.Context, newTask *entity.Task) (int, error)
	UpdateTask(ctx context.Context, id int, updatedTask *entity.Task) (*entity.Task, error)
	DeleteTask(ctx context.Context, id int) error
}

var ErrTaskNotFound = errors.New("task not found")

const dbTimeout = time.Second * 3

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {

	tr := TaskRepository{
		db: db,
	}

	_, err := tr.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INT PRIMARY KEY AUTO_INCREMENT, 
			task_name VARCHAR(45), 
			task_description VARCHAR(45), 
			created_at DATETIME, 
			updated_at DATETIME
	);`)
	if err != nil {
		slog.Error("error creating tasks table", "err", err)
	}

	return &tr
}

func (t *TaskRepository) GetTasksCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var cnt int

	query := "SELECT COUNT(*) FROM tasks"
	err := t.db.QueryRowContext(ctx, query).Scan(&cnt)
	if err != nil {
		slog.Error("Error while getting tasks count", "err", err)
		return 0, err
	}

	return cnt, nil
}

func (t *TaskRepository) GetAllTasks(ctx context.Context) ([]entity.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	result := []entity.Task{}

	rows, err := t.db.QueryContext(ctx, "SELECT * FROM tasks")
	if err != nil {
		slog.Error("Error on query", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fetchedTask entity.Task
		err := rows.Scan(&fetchedTask.Id, &fetchedTask.Name, &fetchedTask.Description, &fetchedTask.CreatedAt, &fetchedTask.UpdatedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, fetchedTask)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TaskRepository) GetTasksPaginated(ctx context.Context, page int, limit int) ([]entity.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	offset := (page - 1) * limit

	result := []entity.Task{}

	rows, err := t.db.QueryContext(ctx, "SELECT * FROM tasks LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task entity.Task
		err := rows.Scan(&task.Id, &task.Name, &task.Description, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, task)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TaskRepository) GetTask(ctx context.Context, taskId int) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	fetchedTask := entity.Task{}

	row := t.db.QueryRowContext(ctx,
		"SELECT id, task_name, task_description, created_at, updated_at FROM tasks WHERE id = ?",
		taskId,
	)

	if err := row.Scan(&fetchedTask.Id, &fetchedTask.Name, &fetchedTask.Description, &fetchedTask.CreatedAt, &fetchedTask.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &fetchedTask, nil
}

func (t *TaskRepository) GetTask_Long(ctx context.Context, taskId int) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	fetchedTask := entity.Task{}

	select {
	// simulates long DB request
	case <-time.After(4 * time.Second):
		row := t.db.QueryRowContext(ctx,
			"SELECT id, task_name, task_description, created_at, updated_at FROM tasks WHERE id = ?",
			taskId,
		)

		if err := row.Scan(&fetchedTask.Id, &fetchedTask.Name, &fetchedTask.Description, &fetchedTask.CreatedAt, &fetchedTask.UpdatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrTaskNotFound
			}
			return nil, err
		}

		return &fetchedTask, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}

	//return &fetchedTask, nil
}

func (t *TaskRepository) CreateTask(ctx context.Context, newTask *entity.Task) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO tasks (task_name, task_description, created_at, updated_at) VALUES(?,?,?,?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, newTask.Name, newTask.Description, newTask.CreatedAt, newTask.UpdatedAt)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return int(lastId), nil
}

func (t *TaskRepository) UpdateTask(ctx context.Context, id int, updatedTask *entity.Task) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	fetchedTask, err := t.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE tasks SET task_name=?, task_description=?, created_at=?, updated_at=? WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, updatedTask.Name, updatedTask.Description, fetchedTask.CreatedAt, time.Now(), id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	updatedTask.CreatedAt = fetchedTask.CreatedAt
	updatedTask.UpdatedAt = time.Now()
	updatedTask.Id = fetchedTask.Id
	//log.Printf("[TaskService] [UpdateTask] Task {%d} updated.", id)

	return updatedTask, nil
}

func (t *TaskRepository) DeleteTask(ctx context.Context, id int) error {

	_, err := t.GetTask(ctx, id)
	if err != nil {
		return err
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM tasks WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	//log.Printf("[TaskService] [DeleteTask] Task with id: {%d} deleted.", id)

	return nil
}
