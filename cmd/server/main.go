package main

import (
	"GoProjects/TaskTracker/internal/store"
	"context"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	db, err := store.NewDB()
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Pool.Close()

	var n int
	err = db.Pool.QueryRow(context.Background(), "select count(*) from users").Scan(&n)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	log.Printf("users in db: %d\n", n)

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			return
		}
	})
	log.Println("server listening on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}
