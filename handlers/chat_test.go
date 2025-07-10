package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"smarapp-api/models"
	"smarapp-api/testutil"
	"smarapp-api/websocket"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestChatHandler_GetChatHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	hub := websocket.NewHub()
	handler := NewChatHandler(hub)

	req := httptest.NewRequest("GET", "/chat/history", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.GET("/chat/history", handler.GetChatHistory)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var messages []models.ChatMessage
	err := json.Unmarshal(w.Body.Bytes(), &messages)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(messages), 1) // We have at least 1 test message

	// Check message structure
	for _, msg := range messages {
		assert.Greater(t, msg.ID, 0)
		assert.Greater(t, msg.UserID, 0)
		assert.NotEmpty(t, msg.Username)
		assert.NotEmpty(t, msg.Message)
		assert.False(t, msg.CreatedAt.IsZero())
	}
}

func TestChatHandler_HandleWebSocket_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	hub := websocket.NewHub()
	handler := NewChatHandler(hub)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing user_id",
			setupContext: func(c *gin.Context) {
				// Don't set user_id
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
		{
			name: "missing username",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				// Don't set username
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Username not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/chat/ws", nil)
			w := httptest.NewRecorder()

			r := gin.New()
			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/chat/ws", handler.HandleWebSocket)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// Note: WebSocket upgrade testing is complex in unit tests due to the need for
// actual WebSocket connections. The authentication logic is tested separately above.

// Note: Integration tests for WebSocket hub functionality would require
// running the hub in a separate goroutine and proper synchronization.
// For now, we test the HTTP endpoints which provide sufficient coverage.
