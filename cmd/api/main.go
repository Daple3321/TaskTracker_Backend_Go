package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gameroll.com/ServerLearn/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	envPath := filepath.Join("..", "..", "configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found at: %s\n", envPath)
		return
	}

	tasksHandler := handlers.NewHandler()
	tasksRouter := tasksHandler.RegisterRoutes()

	router := http.NewServeMux()
	router.Handle("/tasks/", http.StripPrefix("/tasks", tasksRouter))

	// Serve static files
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Starting http server\n")
	err := http.ListenAndServe(os.Getenv("SERVERIP")+":"+os.Getenv("SERVERPORT"), router)
	if err != nil {
		log.Fatal(err)
	}
}
