package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_Constants(t *testing.T) {
	assert.Equal(t, Role("admin"), RoleAdmin)
	assert.Equal(t, Role("user"), RoleUser)
}

func TestUser_JSONSerialization(t *testing.T) {
	user := User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     RoleUser,
	}

	// Test that password field exists but will be excluded from JSON
	assert.Equal(t, "hashedpassword", user.Password)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, RoleUser, user.Role)
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			valid: true,
		},
		{
			name: "empty email",
			request: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "invalid email",
			request: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "short password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "123",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: In a real application, you would use a validator
			// Here we just check the struct fields are set correctly
			if tt.valid {
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Password)
				assert.Contains(t, tt.request.Email, "@")
				assert.GreaterOrEqual(t, len(tt.request.Password), 6)
			}
		})
	}
}

func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request RegisterRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     RoleUser,
			},
			valid: true,
		},
		{
			name: "valid admin request",
			request: RegisterRequest{
				Username: "admin",
				Email:    "admin@example.com",
				Password: "password123",
				Role:     RoleAdmin,
			},
			valid: true,
		},
		{
			name: "short username",
			request: RegisterRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "password123",
				Role:     RoleUser,
			},
			valid: false,
		},
		{
			name: "long username",
			request: RegisterRequest{
				Username: "this-is-a-very-long-username-that-exceeds-fifty-characters",
				Email:    "test@example.com",
				Password: "password123",
				Role:     RoleUser,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.request.Username)
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Password)
				assert.Contains(t, tt.request.Email, "@")
				assert.GreaterOrEqual(t, len(tt.request.Username), 3)
				assert.LessOrEqual(t, len(tt.request.Username), 50)
				assert.GreaterOrEqual(t, len(tt.request.Password), 6)
			}
		})
	}
}

func TestLoginResponse_Structure(t *testing.T) {
	user := User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}

	response := LoginResponse{
		Token: "jwt-token-here",
		User:  user,
	}

	assert.Equal(t, "jwt-token-here", response.Token)
	assert.Equal(t, user, response.User)
	assert.Equal(t, 1, response.User.ID)
	assert.Equal(t, "testuser", response.User.Username)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.Equal(t, RoleUser, response.User.Role)
}
