package websocket

import (
	"encoding/json"
	"log"
	"smarapp-api/database"
	"smarapp-api/models"
	"sync"
	"time"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	dbMutex    sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s (ID: %d)", client.Username, client.UserID)
			
			// Send join message to all clients
			joinMsg := models.WebSocketMessage{
				Type:    models.MessageTypeJoin,
				UserID:  client.UserID,
				Message: client.Username + " joined the chat",
			}
			h.broadcastMessage(joinMsg)
			
			// Send chat history to the new client
			h.sendChatHistory(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected: %s (ID: %d)", client.Username, client.UserID)
				
				// Send leave message to all clients
				leaveMsg := models.WebSocketMessage{
					Type:    models.MessageTypeLeave,
					UserID:  client.UserID,
					Message: client.Username + " left the chat",
				}
				h.broadcastMessage(leaveMsg)
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

func (h *Hub) broadcastMessage(msg models.WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	h.broadcast <- data
}

func (h *Hub) sendChatHistory(client *Client) {
	rows, err := database.DB.Query(`
		SELECT id, user_id, username, message, created_at 
		FROM chat_messages 
		ORDER BY created_at DESC 
		LIMIT 50
	`)
	if err != nil {
		log.Printf("Error fetching chat history: %v", err)
		return
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Message, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning chat message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Reverse to show oldest first
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	historyMsg := models.WebSocketMessage{
		Type: models.MessageTypeHistory,
		Data: messages,
	}

	data, err := json.Marshal(historyMsg)
	if err != nil {
		log.Printf("Error marshaling history: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

func (h *Hub) SaveAndBroadcastMessage(userID int, username, message string) error {
	// Use mutex to prevent concurrent database writes
	h.dbMutex.Lock()
	defer h.dbMutex.Unlock()

	// Save to database
	log.Printf("Attempting to save message: UserID=%d, Username=%s, Message=%s", userID, username, message)
	_, err := database.DB.Exec(
		"INSERT INTO chat_messages (user_id, username, message, created_at) VALUES (?, ?, ?, ?)",
		userID, username, message, time.Now(),
	)
	if err != nil {
		log.Printf("Error saving message to database: %v", err)
		return err
	}
	log.Printf("Message saved successfully to database")

	// Broadcast to all clients
	chatMsg := models.WebSocketMessage{
		Type:    models.MessageTypeChat,
		UserID:  userID,
		Message: message,
		Data: models.ChatMessage{
			UserID:    userID,
			Username:  username,
			Message:   message,
			CreatedAt: time.Now(),
		},
	}

	h.broadcastMessage(chatMsg)
	return nil
}
