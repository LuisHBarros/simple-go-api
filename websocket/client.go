package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"smarapp-api/models"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	UserID   int
	Username string
}

func (c *Client) readPump() {
	defer func() {
		log.Printf("Client %s (ID: %d) readPump exiting", c.Username, c.UserID)
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	log.Printf("Starting readPump for client %s (ID: %d)", c.Username, c.UserID)
	for {
		log.Printf("Waiting for message from client %s (ID: %d)", c.Username, c.UserID)
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Printf("WebSocket connection closed: %v", err)
			}
			break
		}

		log.Printf("Raw message received from client %s (ID: %d): %s", c.Username, c.UserID, string(messageBytes))

		var msg models.SendMessageRequest
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		log.Printf("Parsed message from client %s (ID: %d): %s", c.Username, c.UserID, msg.Message)

		// Save and broadcast the message
		if err := c.hub.SaveAndBroadcastMessage(c.UserID, c.Username, msg.Message); err != nil {
			log.Printf("Error saving message: %v", err)
		} else {
			log.Printf("Message processed successfully, continuing loop...")
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, userID int, username string) {
	log.Printf("ServeWS called for user %s (ID: %d)", username, userID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	log.Printf("WebSocket upgrade successful for user %s (ID: %d)", username, userID)

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
	}

	log.Printf("Registering client %s (ID: %d)", username, userID)
	client.hub.register <- client

	log.Printf("Starting goroutines for client %s (ID: %d)", username, userID)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("writePump panic for client %s (ID: %d): %v", username, userID, r)
			}
		}()
		client.writePump()
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("readPump panic for client %s (ID: %d): %v", username, userID, r)
			}
		}()
		client.readPump()
	}()

	log.Printf("Goroutines started for client %s (ID: %d)", username, userID)
}
