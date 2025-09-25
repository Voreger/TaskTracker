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
	r.Use(handlers.Logger)
	db, err := store.NewDB()
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Pool.Close()

	userStore := store.NewUserStore(db.Pool)
	handlers.RegisterUserRoutes(r, userStore)
	handlers.RegisterAuthRoutes(r, userStore)

	r.Group(func(pr chi.Router) {
		pr.Use(handlers.AuthMiddleware)
		handlers.RegisterTaskRoutes(pr, store.NewTaskStore(db.Pool))
	})

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
