package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFormatValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test struct with validation tags
	type TestRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Age      int    `json:"age" binding:"required,gte=18"`
	}

	tests := []struct {
		name           string
		requestBody    string
		expectedErrors int
		expectedFields []string
	}{
		{
			name:           "missing required fields",
			requestBody:    `{}`,
			expectedErrors: 3,
			expectedFields: []string{"email", "password", "age"},
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email", "password": "123456", "age": 25}`,
			expectedErrors: 1,
			expectedFields: []string{"email"},
		},
		{
			name:           "password too short",
			requestBody:    `{"email": "test@example.com", "password": "123", "age": 25}`,
			expectedErrors: 1,
			expectedFields: []string{"password"},
		},
		{
			name:           "age too low",
			requestBody:    `{"email": "test@example.com", "password": "123456", "age": 16}`,
			expectedErrors: 1,
			expectedFields: []string{"age"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate request
			c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := c.ShouldBindJSON(&req)

			if err != nil {
				errors := FormatValidationErrors(err)
				assert.Len(t, errors, tt.expectedErrors)

				// Check that expected fields are present
				for _, expectedField := range tt.expectedFields {
					found := false
					for _, apiErr := range errors {
						if apiErr.Field == expectedField {
							found = true
							assert.Equal(t, ErrCodeValidationFailed, apiErr.Code)
							assert.NotEmpty(t, apiErr.Message)
							break
						}
					}
					assert.True(t, found, "Expected field %s not found in errors", expectedField)
				}
			}
		})
	}
}

func TestAPIErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		error          *APIError
		expectedStatus int
		expectedCode   ErrorCode
	}{
		{
			name:           "unauthorized error",
			error:          Unauthorized("Authentication required"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   ErrCodeUnauthorized,
		},
		{
			name:           "not found error",
			error:          NotFound("User"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrCodeNotFound,
		},
		{
			name:           "validation error",
			error:          ValidationError("Invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrCodeValidationFailed,
		},
		{
			name:           "internal error",
			error:          InternalError("Something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RespondWithAPIError(c, tt.error)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCode, response.Error.Code)
			assert.NotEmpty(t, response.Error.Message)
		})
	}
}

func TestBusinessLogicErrors(t *testing.T) {
	tests := []struct {
		name     string
		error    *APIError
		expected ErrorCode
	}{
		{
			name:     "insufficient stock",
			error:    InsufficientStock("Product A", 5, 3),
			expected: ErrCodeInsufficientStock,
		},
		{
			name:     "product not found",
			error:    ProductNotFound(123),
			expected: ErrCodeNotFound,
		},
		{
			name:     "user already exists",
			error:    UserAlreadyExists("email", "test@example.com"),
			expected: ErrCodeAlreadyExists,
		},
		{
			name:     "invalid credentials",
			error:    InvalidCredentials(),
			expected: ErrCodeInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.error.Code)
			assert.NotEmpty(t, tt.error.Message)
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		code           ErrorCode
		expectedStatus int
	}{
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeValidationFailed, http.StatusBadRequest},
		{ErrCodeAlreadyExists, http.StatusConflict},
		{ErrCodeInternalError, http.StatusInternalServerError},
		{ErrCodeDatabaseError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			status := GetHTTPStatus(tt.code)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestValidationHelpers(t *testing.T) {
	t.Run("ValidateRole", func(t *testing.T) {
		assert.Nil(t, ValidateRole("user"))
		assert.Nil(t, ValidateRole("admin"))
		assert.NotNil(t, ValidateRole("invalid"))
	})

	t.Run("ValidatePositiveQuantity", func(t *testing.T) {
		assert.Nil(t, ValidatePositiveQuantity(5, "quantity"))
		assert.NotNil(t, ValidatePositiveQuantity(0, "quantity"))
		assert.NotNil(t, ValidatePositiveQuantity(-1, "quantity"))
	})

	t.Run("ValidateEmail", func(t *testing.T) {
		assert.Nil(t, ValidateEmail("test@example.com"))
		assert.NotNil(t, ValidateEmail("invalid-email"))
		assert.NotNil(t, ValidateEmail(""))
	})

	t.Run("ValidatePassword", func(t *testing.T) {
		assert.Nil(t, ValidatePassword("password123"))
		assert.NotNil(t, ValidatePassword("123"))
		assert.NotNil(t, ValidatePassword(""))
	})
}
