package main

import (
	"GoProjects/TaskTracker/internal/handlers"
	"GoProjects/TaskTracker/internal/store"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	db, err := store.NewDB()
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Pool.Close()

	taskStore := store.NewTaskStore(db.Pool)
	handlers.RegisterTaskRoutes(r, taskStore)

	userStore := store.NewUserStore(db.Pool)
	handlers.RegisterUserRoutes(r, userStore)

	log.Println("server listening on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}
