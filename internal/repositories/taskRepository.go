package repositories

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/Daple3321/TaskTracker/internal/entity"
)

type Repository interface {
	GetTasksCount() (int, error)
	GetAllTasks() ([]entity.Task, error)
	GetTasksPaginated(offset int, limit int) ([]entity.Task, error)
	GetTask(taskId int) (*entity.Task, error)
	CreateTask(newTask *entity.Task) (int, error)
	UpdateTask(id int, updatedTask *entity.Task) (*entity.Task, error)
	DeleteTask(id int) error
}

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {

	tr := TaskRepository{
		db: db,
	}

	return &tr
}

func (t *TaskRepository) GetTasksCount() (int, error) {

	var cnt int

	query := "SELECT COUNT(*) FROM tasks"
	err := t.db.QueryRow(query).Scan(&cnt)
	if err != nil {
		slog.Error("Error while getting tasks count", "err", err)
		return 0, err
	}

	return cnt, nil
}

func (t *TaskRepository) GetAllTasks() ([]entity.Task, error) {

	result := []entity.Task{}

	rows, err := t.db.Query("SELECT * FROM tasks")
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

func (t *TaskRepository) GetTasksPaginated(page int, limit int) ([]entity.Task, error) {

	offset := (page - 1) * limit

	result := []entity.Task{}

	rows, err := t.db.Query("SELECT * FROM tasks LIMIT ? OFFSET ?", limit, offset)
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

func (t *TaskRepository) GetTask(taskId int) (*entity.Task, error) {

	fetchedTask := entity.Task{}

	row := t.db.QueryRow(
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

func (t *TaskRepository) CreateTask(newTask *entity.Task) (int, error) {

	tx, err := t.db.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tasks (task_name, task_description, created_at, updated_at) VALUES(?,?,?,?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(newTask.Name, newTask.Description, newTask.CreatedAt, newTask.UpdatedAt)
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

func (t *TaskRepository) UpdateTask(id int, updatedTask *entity.Task) (*entity.Task, error) {

	fetchedTask, err := t.GetTask(id)
	if err != nil {
		return nil, err
	}

	tx, err := t.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE tasks SET task_name=?, task_description=?, created_at=?, updated_at=? WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(updatedTask.Name, updatedTask.Description, fetchedTask.CreatedAt, time.Now(), id)
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

func (t *TaskRepository) DeleteTask(id int) error {

	_, err := t.GetTask(id)
	if err != nil {
		return err
	}

	tx, err := t.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM tasks WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
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
