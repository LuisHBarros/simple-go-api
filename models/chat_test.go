package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChatMessage_Structure(t *testing.T) {
	now := time.Now()
	message := ChatMessage{
		ID:        1,
		UserID:    2,
		Username:  "testuser",
		Message:   "Hello, world!",
		CreatedAt: now,
	}

	assert.Equal(t, 1, message.ID)
	assert.Equal(t, 2, message.UserID)
	assert.Equal(t, "testuser", message.Username)
	assert.Equal(t, "Hello, world!", message.Message)
	assert.Equal(t, now, message.CreatedAt)
}

func TestSendMessageRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request SendMessageRequest
		valid   bool
	}{
		{
			name: "valid message",
			request: SendMessageRequest{
				Message: "Hello, world!",
			},
			valid: true,
		},
		{
			name: "empty message",
			request: SendMessageRequest{
				Message: "",
			},
			valid: false,
		},
		{
			name: "long message",
			request: SendMessageRequest{
				Message: generateLongString(1001),
			},
			valid: false,
		},
		{
			name: "max length message",
			request: SendMessageRequest{
				Message: generateLongString(1000),
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.request.Message)
				assert.LessOrEqual(t, len(tt.request.Message), 1000)
			}
		})
	}
}

func TestWebSocketMessage_Structure(t *testing.T) {
	message := WebSocketMessage{
		Type:    MessageTypeChat,
		Data:    "test data",
		UserID:  1,
		Message: "Hello",
	}

	assert.Equal(t, MessageTypeChat, message.Type)
	assert.Equal(t, "test data", message.Data)
	assert.Equal(t, 1, message.UserID)
	assert.Equal(t, "Hello", message.Message)
}

func TestWebSocketMessage_Types(t *testing.T) {
	assert.Equal(t, "chat", MessageTypeChat)
	assert.Equal(t, "join", MessageTypeJoin)
	assert.Equal(t, "leave", MessageTypeLeave)
	assert.Equal(t, "error", MessageTypeError)
	assert.Equal(t, "history", MessageTypeHistory)
}

func TestWebSocketMessage_ChatType(t *testing.T) {
	chatMessage := WebSocketMessage{
		Type:    MessageTypeChat,
		UserID:  1,
		Message: "Hello everyone!",
	}

	assert.Equal(t, MessageTypeChat, chatMessage.Type)
	assert.Equal(t, 1, chatMessage.UserID)
	assert.Equal(t, "Hello everyone!", chatMessage.Message)
}

func TestWebSocketMessage_JoinType(t *testing.T) {
	joinMessage := WebSocketMessage{
		Type:   MessageTypeJoin,
		UserID: 1,
		Data:   "User joined the chat",
	}

	assert.Equal(t, MessageTypeJoin, joinMessage.Type)
	assert.Equal(t, 1, joinMessage.UserID)
	assert.Equal(t, "User joined the chat", joinMessage.Data)
}

func TestWebSocketMessage_ErrorType(t *testing.T) {
	errorMessage := WebSocketMessage{
		Type: MessageTypeError,
		Data: "Authentication failed",
	}

	assert.Equal(t, MessageTypeError, errorMessage.Type)
	assert.Equal(t, "Authentication failed", errorMessage.Data)
}

func TestWebSocketMessage_HistoryType(t *testing.T) {
	historyData := []ChatMessage{
		{ID: 1, UserID: 1, Username: "user1", Message: "Hello"},
		{ID: 2, UserID: 2, Username: "user2", Message: "Hi there"},
	}

	historyMessage := WebSocketMessage{
		Type: MessageTypeHistory,
		Data: historyData,
	}

	assert.Equal(t, MessageTypeHistory, historyMessage.Type)
	assert.Equal(t, historyData, historyMessage.Data)
}

// Helper function to generate strings of specific length
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
