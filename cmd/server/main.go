package main

import (
	"GoProjects/TaskTracker/internal/cache"
	_ "GoProjects/TaskTracker/internal/docs"
	"GoProjects/TaskTracker/internal/handlers"
	"GoProjects/TaskTracker/internal/queue"
	"GoProjects/TaskTracker/internal/realtime"
	"GoProjects/TaskTracker/internal/store"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	r := chi.NewRouter()
	r.Use(handlers.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Mount("/swagger/", httpSwagger.WrapHandler)
	db, err := store.NewDB()
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Pool.Close()

	hub := realtime.NewHub()
	go hub.Run()

	handlers.RegisterWSRoutes(r, hub)

	broker, err := queue.NewBroker("amqp://guest:guest@rabbitmq:5672/", "tasks")
	if err != nil {
		log.Fatalf("RabbitMQ error: %v", err)
	}
	defer broker.Close()

	redisCache := cache.NewRedisCache("redis:6379")
	defer redisCache.Close()

	userStore := store.NewUserStore(db.Pool)
	handlers.RegisterUserRoutes(r, userStore)
	handlers.RegisterAuthRoutes(r, userStore)

	r.Group(func(pr chi.Router) {
		pr.Use(handlers.AuthMiddleware)
		handlers.RegisterTaskRoutes(pr, store.NewTaskStore(db.Pool), hub, broker, redisCache)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on %s", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
