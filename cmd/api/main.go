package main

import (
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/Daple3321/ServerLearn/internal/handlers"
	"github.com/Daple3321/ServerLearn/internal/middleware"
	"github.com/joho/godotenv"
)

func main() {

	workDir, _ := os.Getwd()
	logPath := path.Join(workDir, "server.log")
	//os.WriteFile(logPath, []byte{}, os.ModeAppend)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "err", err)
		return
	}
	defer logFile.Close()

	// TODO: Make logs to file
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &opts))
	slog.SetDefault(logger)

	envPath := filepath.Join("..", "..", "configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Error("No .env file found at:", "envPath", envPath)
		return
	}

	tasksHandler := handlers.NewHandler()
	tasksRouter := tasksHandler.RegisterRoutes()

	authHandler := middleware.NewAuthHandler()
	authRouter := authHandler.RegisterRoutes()

	router := http.NewServeMux()
	router.Handle("/tasks/", http.StripPrefix("/tasks", tasksRouter))

	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))

	// Serve static files
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	slog.Info("Listening on:", "ip", os.Getenv("SERVERIP"), "port", os.Getenv("SERVERPORT"))
	err = http.ListenAndServe(os.Getenv("SERVERIP")+":"+os.Getenv("SERVERPORT"), router)
	if err != nil {
		slog.Error("Error starting http server", "err", err)
	}
}
