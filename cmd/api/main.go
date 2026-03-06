package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/Daple3321/TaskTracker/internal/handlers"
	"github.com/Daple3321/TaskTracker/internal/middleware"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	logFile, err := SetupLogger()
	if err != nil {
		return
	}
	defer logFile.Close()

	envPath := filepath.Join("..", "..", "configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Error("No .env file found at:", "envPath", envPath)
		return
	}

	db, err := SetupDB()
	if err != nil {
		return
	}

	tasksHandler := handlers.NewHandler(db)
	tasksRouter := tasksHandler.RegisterRoutes()

	authHandler := middleware.NewAuthHandler()
	authRouter := authHandler.RegisterRoutes()

	router := http.NewServeMux()
	router.Handle("/tasks/", http.StripPrefix("/tasks", tasksRouter))

	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))

	slog.Info("Listening on:", "ip", os.Getenv("SERVERIP"), "port", os.Getenv("SERVERPORT"))
	err = http.ListenAndServe(os.Getenv("SERVERIP")+":"+os.Getenv("SERVERPORT"), router)
	if err != nil {
		slog.Error("Error starting http server", "err", err)
	}
}

func SetupDB() (*sql.DB, error) {
	newDb, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/tasks?parseTime=true",
			os.Getenv("TASKDB_USERNAME"),
			os.Getenv("TASKDB_PASSWORD"),
			os.Getenv("SERVERIP")))
	if err != nil {
		slog.Error("Error opening database", "err", err)
		return nil, err
	}

	pingErr := newDb.Ping()
	if pingErr != nil {
		slog.Error("Error while pinging DB", "err", pingErr)
		return nil, pingErr
	}

	return newDb, nil
}

func SetupLogger() (*os.File, error) {
	workDir, _ := os.Getwd()
	logPath := path.Join(workDir, "server.log")
	//os.WriteFile(logPath, []byte{}, os.ModeAppend)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "err", err)
		return nil, err
	}

	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &opts))
	slog.SetDefault(logger)

	return logFile, nil
}
