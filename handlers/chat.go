package handlers

import (
	"net/http"
	"smarapp-api/database"
	"smarapp-api/models"
	"smarapp-api/websocket"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	Hub *websocket.Hub
}

func NewChatHandler(hub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		Hub: hub,
	}
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found"})
		return
	}

	websocket.ServeWS(h.Hub, c.Writer, c.Request, userID.(int), username.(string))
}

func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, user_id, username, message, created_at 
		FROM chat_messages 
		ORDER BY created_at DESC 
		LIMIT 100
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Message, &msg.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan message"})
			return
		}
		messages = append(messages, msg)
	}

	// Reverse to show oldest first
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	c.JSON(http.StatusOK, messages)
}
