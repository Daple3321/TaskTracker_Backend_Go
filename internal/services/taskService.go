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
		log.Fatalf("Error opening database, %s", err)
	}

	pingErr := newDb.Ping()
	if pingErr != nil {
		log.Printf("Error while pinging, %s", pingErr)
	}

	ts := TaskService{
		db: newDb,
	}

	newId := ts.AddTask(&entity.Task{
		Name:        "Bullshit",
		Description: "...",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	ts.GetTask(int(newId))

	return &ts
}

func (t *TaskService) GetTasksCount() int {

	var cnt int

	query := "SELECT COUNT(*) FROM tasks.tasks"
	err := t.db.QueryRow(query).Scan(&cnt)
	if err != nil {
		log.Fatal(err)
	}

	return cnt
}

func (t *TaskService) GetTask(taskId int) *entity.Task {

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

		log.Println(fetchedTask)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return &fetchedTask
}

func (t *TaskService) AddTask(newTask *entity.Task) int64 {

	res, err := t.db.Exec("INSERT INTO tasks (task_name, task_description, created_at, updated_at) VALUES(?,?,?,?)",
		newTask.Name,
		newTask.Description,
		newTask.CreatedAt,
		newTask.UpdatedAt)
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

	log.Printf("[TaskService] [AddTask] Task {%d} added. Rows affected: %d", lastId, rowCnt)

	return lastId
}
