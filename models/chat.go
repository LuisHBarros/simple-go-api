package models

import (
	"time"
)

type ChatMessage struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SendMessageRequest struct {
	Message string `json:"message" binding:"required,min=1,max=1000"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	UserID  int         `json:"user_id,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WebSocket message types
const (
	MessageTypeChat     = "chat"
	MessageTypeJoin     = "join"
	MessageTypeLeave    = "leave"
	MessageTypeError    = "error"
	MessageTypeHistory  = "history"
)
