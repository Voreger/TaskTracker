package main

import (
	"GoProjects/TaskTracker/internal/cache"
	_ "GoProjects/TaskTracker/internal/docs"
	"GoProjects/TaskTracker/internal/handlers"
	"GoProjects/TaskTracker/internal/logger"
	"GoProjects/TaskTracker/internal/queue"
	"GoProjects/TaskTracker/internal/realtime"
	"GoProjects/TaskTracker/internal/store"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Init()
	defer logger.Sync()

	r := chi.NewRouter()
	r.Use(handlers.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Handle("/metrics", promhttp.Handler())
	r.Mount("/swagger/", httpSwagger.WrapHandler)
	db, err := store.NewDB()
	if err != nil {
		logger.Log.Fatal("db error", zap.Error(err))
	}
	defer db.Pool.Close()

	hub := realtime.NewHub()
	go hub.Run(ctx)

	handlers.RegisterWSRoutes(r, hub)

	broker, err := queue.NewBroker("amqp://guest:guest@rabbitmq:5672/", "tasks")
	if err != nil {
		logger.Log.Fatal("RabbitMQ error", zap.Error(err))
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
		logger.Log.Info("server listening", zap.String("addr", srv.Addr))
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server error", zap.Error(err))
		}
	}()

	<-stop
	logger.Log.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.Log.Fatal("server shutdown error", zap.Error(err))
	}

	logger.Log.Info("Server gracefully stopped")

}
