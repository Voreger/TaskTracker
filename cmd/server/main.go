package main

import (
	_ "GoProjects/TaskTracker/internal/docs"
	"GoProjects/TaskTracker/internal/handlers"
	"GoProjects/TaskTracker/internal/realtime"
	"GoProjects/TaskTracker/internal/store"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(handlers.Logger)
	r.Mount("/swagger/", httpSwagger.WrapHandler)
	db, err := store.NewDB()
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Pool.Close()

	hub := realtime.NewHub()
	go hub.Run()

	handlers.RegisterWSRoutes(r, hub)

	userStore := store.NewUserStore(db.Pool)
	handlers.RegisterUserRoutes(r, userStore)
	handlers.RegisterAuthRoutes(r, userStore)

	r.Group(func(pr chi.Router) {
		pr.Use(handlers.AuthMiddleware)
		handlers.RegisterTaskRoutes(pr, store.NewTaskStore(db.Pool), hub)
	})

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
