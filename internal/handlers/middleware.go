package handlers

import (
	"GoProjects/TaskTracker/internal/auth"
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(status int) {
	if rw.wroteHeader {
		return
	}
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
	rw.wroteHeader = true
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" || websocket.IsWebSocketUpgrade(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		log.Printf("[%s] %s %d %s", r.Method, r.RequestURI, rw.status, time.Since(start))
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ParseToken(parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
