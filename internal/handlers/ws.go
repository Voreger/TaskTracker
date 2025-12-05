package handlers

import (
	"GoProjects/TaskTracker/internal/logger"
	"GoProjects/TaskTracker/internal/realtime"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	Hub *realtime.Hub
}

func RegisterWSRoutes(r chi.Router, hub *realtime.Hub) {
	h := &WSHandler{Hub: hub}
	r.Get("/ws", h.HandleWS)
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("ws upgrade failed", zap.Error(err))
		http.Error(w, "could not upgrade", http.StatusInternalServerError)
		return
	}

	client := realtime.NewClient(h.Hub, conn)
	h.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
