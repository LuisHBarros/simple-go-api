package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"smarapp-api/models"
	"smarapp-api/testutil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	handler := NewAuthHandler("test-secret")

	tests := []struct {
		name           string
		requestBody    models.RegisterRequest
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid user registration",
			requestBody: models.RegisterRequest{
				Username: "newuser",
				Email:    "newuser@test.com",
				Password: "password123",
				Role:     models.RoleUser,
			},
			expectedStatus: http.StatusCreated,
			expectToken:    true,
		},
		{
			name: "valid admin registration",
			requestBody: models.RegisterRequest{
				Username: "newadmin",
				Email:    "newadmin@test.com",
				Password: "password123",
				Role:     models.RoleAdmin,
			},
			expectedStatus: http.StatusCreated,
			expectToken:    true,
		},
		{
			name: "registration without role (defaults to user)",
			requestBody: models.RegisterRequest{
				Username: "defaultuser",
				Email:    "defaultuser@test.com",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			expectToken:    true,
		},
		{
			name: "invalid email",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "short password",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@test.com",
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "invalid role",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@test.com",
				Password: "password123",
				Role:     "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create router and handle request
			r := gin.New()
			r.POST("/register", handler.Register)
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectToken {
				var response models.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, tt.requestBody.Username, response.User.Username)
				assert.Equal(t, tt.requestBody.Email, response.User.Email)
				
				// Check role default
				expectedRole := tt.requestBody.Role
				if expectedRole == "" {
					expectedRole = models.RoleUser
				}
				assert.Equal(t, expectedRole, response.User.Role)
			}
		})
	}
}

func TestAuthHandler_Register_DuplicateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewAuthHandler("test-secret")

	tests := []struct {
		name        string
		requestBody models.RegisterRequest
	}{
		{
			name: "duplicate email",
			requestBody: models.RegisterRequest{
				Username: "newuser",
				Email:    "admin@test.com", // This email already exists
				Password: "password123",
			},
		},
		{
			name: "duplicate username",
			requestBody: models.RegisterRequest{
				Username: "admin", // This username already exists
				Email:    "newemail@test.com",
				Password: "password123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r := gin.New()
			r.POST("/register", handler.Register)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusConflict, w.Code)
			assert.Contains(t, w.Body.String(), "already exists")
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	handler := NewAuthHandler("test-secret")

	// First register a user
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@test.com",
		Password: "password123",
		Role:     models.RoleUser,
	}
	jsonBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/register", handler.Register)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid login",
			requestBody: models.LoginRequest{
				Email:    "test@test.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "invalid email",
			requestBody: models.LoginRequest{
				Email:    "nonexistent@test.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "invalid password",
			requestBody: models.LoginRequest{
				Email:    "test@test.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "malformed email",
			requestBody: models.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r := gin.New()
			r.POST("/login", handler.Login)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectToken {
				var response models.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "testuser", response.User.Username)
				assert.Equal(t, "test@test.com", response.User.Email)
				assert.Equal(t, models.RoleUser, response.User.Role)
			}
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewAuthHandler("test-secret")

	tests := []struct {
		name           string
		userID         int
		expectedStatus int
	}{
		{
			name:           "valid user profile",
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent user",
			userID:         999,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/profile", nil)
			w := httptest.NewRecorder()

			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})
			r.GET("/profile", handler.GetProfile)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var user models.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, user.ID)
				assert.NotEmpty(t, user.Username)
				assert.NotEmpty(t, user.Email)
			}
		})
	}
}
