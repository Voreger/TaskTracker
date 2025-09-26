package realtime

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client
	unregister chan *Client
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
func (h *Hub) Broadcast(message Message) {
	data, _ := json.Marshal(message)
	h.broadcast <- data
}
