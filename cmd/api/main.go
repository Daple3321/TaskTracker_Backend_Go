package main

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logFile, err := SetupLogger()
	if err != nil {
		return
	}
	defer logFile.Close()

	envPath := filepath.Join("..", "..", "configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Error("no .env file found at:", "envPath", envPath)
		return
	}

	if err := ValidateEnvVars(); err != nil {
		slog.Error("error validating env vars", "err", err)
		return
	}

	db, err := SetupDB()
	if err != nil {
		return
	}

	go middleware.LimitTimeoutRoutine(ctx)

	tasksHandler := handlers.NewTaskHandler(db)
	tasksRouter := tasksHandler.RegisterRoutes()

	usersHandler := handlers.NewUsersHandler(db)
	authRouter := usersHandler.RegisterRoutes()

	router := http.NewServeMux()
	router.Handle("/tasks/", http.StripPrefix("/tasks", tasksRouter))

	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))

	handler := corsMiddleware(router)

	slog.Info("Listening on:", "ip", os.Getenv("SERVERIP"), "port", os.Getenv("SERVERPORT"))
	err = http.ListenAndServe(os.Getenv("SERVERIP")+":"+os.Getenv("SERVERPORT"), handler)
	if err != nil {
		slog.Error("error starting http server", "err", err)
	}
}

func ValidateEnvVars() error {
	vars := []string{
		"SERVERPORT",
		"SERVERIP",
		"TASKDB_USERNAME",
		"TASKDB_PASSWORD",
		"JWT_SECRET_KEY",
	}

	for _, v := range vars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("env var %s not set", v)
		}
	}

	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetupDB() (*sql.DB, error) {
	newDb, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/tasks?parseTime=true",
			os.Getenv("TASKDB_USERNAME"),
			os.Getenv("TASKDB_PASSWORD"),
			os.Getenv("SERVERIP")))
	if err != nil {
		slog.Error("error opening database", "err", err)
		return nil, err
	}

	pingErr := newDb.Ping()
	if pingErr != nil {
		slog.Error("error while pinging DB", "err", pingErr)
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
		slog.Error("failed to open log file", "err", err)
		return nil, err
	}

	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &opts))
	slog.SetDefault(logger)

	return logFile, nil
}
