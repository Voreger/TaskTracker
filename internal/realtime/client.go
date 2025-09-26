package realtime

import "github.com/gorilla/websocket"

func NewClient(h *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
