package services

import (
	"fmt"
	"log"
	"os"
	"time"

	"gameroll.com/ServerLearn/internal/entity"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const TasksTable string = "tasks.tasks"

type TaskService struct {
	db *sql.DB
}

func NewTaskService() *TaskService {

	newDb, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/tasks?parseTime=true",
			os.Getenv("TASKDB_USERNAME"),
			os.Getenv("TASKDB_PASSWORD"),
			os.Getenv("SERVERIP")))
	if err != nil {
		log.Fatalf("[TaskService][New] Error opening database, %s", err)
	}

	pingErr := newDb.Ping()
	if pingErr != nil {
		log.Printf("Error while pinging, %s", pingErr)
	}

	ts := TaskService{
		db: newDb,
	}

	newId := ts.AddTask(&entity.Task{
		Name:        "NEW Bullshit",
		Description: "...asdasd23#@#@34",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	ts.GetTask(int(newId))

	return &ts
}

func (t *TaskService) GetTasksCount() int {

	var cnt int

	query := "SELECT COUNT(*) FROM tasks"
	err := t.db.QueryRow(query).Scan(&cnt)
	if err != nil {
		log.Fatal(err)
	}

	return cnt
}

func (t *TaskService) GetTasks() []entity.Task {

	result := []entity.Task{}

	rows, err := t.db.Query("SELECT * FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var fetchedTask entity.Task
		err := rows.Scan(&fetchedTask.Id, &fetchedTask.Name, &fetchedTask.Description, &fetchedTask.CreatedAt, &fetchedTask.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}

		result = append(result, fetchedTask)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func (t *TaskService) GetTask(taskId int) (*entity.Task, error) {

	fetchedTask := entity.Task{}

	rows, err := t.db.Query("select * from tasks where id = ?", taskId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&fetchedTask.Id, &fetchedTask.Name, &fetchedTask.Description, &fetchedTask.CreatedAt, &fetchedTask.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}

	}

	// bad check.
	if fetchedTask.Name == "" {
		return nil, fmt.Errorf("No task with id {%d} found.", taskId)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return &fetchedTask, nil
}

func (t *TaskService) AddTask(newTask *entity.Task) int {

	tx, err := t.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tasks (task_name, task_description, created_at, updated_at) VALUES(?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	res, err := stmt.Exec(newTask.Name, newTask.Description, newTask.CreatedAt, newTask.UpdatedAt)
	if err != nil {
		log.Fatal(err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[TaskService] [AddTask] Task {%d} added. Rows affected: %d", lastId, rowCnt)

	return int(lastId)
}

func (t *TaskService) UpdateTask(id int, updatedTask *entity.Task) (*entity.Task, error) {

	fetchedTask, err := t.GetTask(id)
	if err != nil {
		return nil, err
	}

	tx, err := t.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE tasks SET task_name=?, task_description=?, created_at=?, updated_at=? WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(updatedTask.Name, updatedTask.Description, fetchedTask.CreatedAt, time.Now(), id)
	if err != nil {
		log.Fatal(err)
	}

	// rowCnt, err := res.RowsAffected()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if rowCnt == 0 {
	// 	return nil, fmt.Errorf("Error updating task with id: %d. Wrong id?", id)
	// }

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	updatedTask.CreatedAt = fetchedTask.CreatedAt
	updatedTask.UpdatedAt = time.Now()
	updatedTask.Id = fetchedTask.Id
	log.Printf("[TaskService] [UpdateTask] Task {%d} updated.", id)

	return updatedTask, nil
}

func (t *TaskService) DeleteTask(id int) error {

	_, err := t.GetTask(id)
	if err != nil {
		return err
	}

	tx, err := t.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM tasks WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[TaskService] [DeleteTask] Task with id: {%d} deleted.", id)

	return nil
}
