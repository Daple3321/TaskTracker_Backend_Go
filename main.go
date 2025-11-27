package main

import (
	"log"
	"net/http"
	"os"

	"gameroll.com/ServerLearn/services/tasks"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	tasksHandler := tasks.NewHandler()
	tasksRouter := tasksHandler.RegisterRoutes()

	router := http.NewServeMux()
	router.Handle("/tasks/", http.StripPrefix("/tasks", tasksRouter))

	err := http.ListenAndServe("localhost:"+os.Getenv("SERVERPORT"), router)
	if err != nil {
		log.Fatal(err)
	}
}
