package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Authentication errors
	ErrCodeInvalidCredentials    ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeTokenExpired        ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid        ErrorCode = "TOKEN_INVALID"
	ErrCodeMissingAuth         ErrorCode = "MISSING_AUTH"
	ErrCodeInvalidRole         ErrorCode = "INVALID_ROLE"

	// Validation errors
	ErrCodeValidationFailed    ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField        ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat       ErrorCode = "INVALID_FORMAT"

	// Resource errors
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict           ErrorCode = "CONFLICT"

	// Business logic errors
	ErrCodeInsufficientStock  ErrorCode = "INSUFFICIENT_STOCK"
	ErrCodeInvalidQuantity    ErrorCode = "INVALID_QUANTITY"
	ErrCodeOrderNotAllowed    ErrorCode = "ORDER_NOT_ALLOWED"

	// System errors
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// APIError represents a structured API error
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Field   string    `json:"field,omitempty"`
}

// ErrorResponse represents the complete error response
type ErrorResponse struct {
	Error  APIError    `json:"error"`
	Errors []APIError  `json:"errors,omitempty"`
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, details ...string) *APIError {
	err := &APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// NewFieldError creates a new field-specific API error
func NewFieldError(code ErrorCode, field, message string, details ...string) *APIError {
	err := &APIError{
		Code:    code,
		Field:   field,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// RespondWithError sends a single error response
func RespondWithError(c *gin.Context, statusCode int, err *APIError) {
	response := ErrorResponse{Error: *err}
	c.JSON(statusCode, response)
}

// RespondWithErrors sends multiple errors response
func RespondWithErrors(c *gin.Context, statusCode int, errors []APIError) {
	if len(errors) == 1 {
		RespondWithError(c, statusCode, &errors[0])
		return
	}
	
	response := ErrorResponse{
		Error:  errors[0], // Primary error
		Errors: errors,    // All errors
	}
	c.JSON(statusCode, response)
}

// Common error constructors
func Unauthorized(message string, details ...string) *APIError {
	if message == "" {
		message = "Authentication required"
	}
	return NewAPIError(ErrCodeUnauthorized, message, details...)
}

func Forbidden(message string, details ...string) *APIError {
	if message == "" {
		message = "Access denied"
	}
	return NewAPIError(ErrCodeForbidden, message, details...)
}

func NotFound(resource string, details ...string) *APIError {
	message := fmt.Sprintf("%s not found", resource)
	return NewAPIError(ErrCodeNotFound, message, details...)
}

func AlreadyExists(resource string, details ...string) *APIError {
	message := fmt.Sprintf("%s already exists", resource)
	return NewAPIError(ErrCodeAlreadyExists, message, details...)
}

func InternalError(message string, details ...string) *APIError {
	if message == "" {
		message = "An internal error occurred"
	}
	return NewAPIError(ErrCodeInternalError, message, details...)
}

func DatabaseError(operation string, details ...string) *APIError {
	message := fmt.Sprintf("Database error during %s", operation)
	return NewAPIError(ErrCodeDatabaseError, message, details...)
}

func ValidationError(message string, details ...string) *APIError {
	if message == "" {
		message = "Validation failed"
	}
	return NewAPIError(ErrCodeValidationFailed, message, details...)
}

func InvalidInput(field, message string, details ...string) *APIError {
	return NewFieldError(ErrCodeInvalidInput, field, message, details...)
}

// FormatValidationErrors converts Gin validation errors to structured API errors
func FormatValidationErrors(err error) []APIError {
	var errors []APIError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			field := strings.ToLower(fieldError.Field())
			message := getValidationMessage(fieldError)
			
			apiError := NewFieldError(ErrCodeValidationFailed, field, message)
			errors = append(errors, *apiError)
		}
	} else {
		// Fallback for other binding errors
		errors = append(errors, *ValidationError("Invalid request format", err.Error()))
	}

	return errors
}

// getValidationMessage returns a user-friendly message for validation errors
func getValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at most %s characters long", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// HTTP status code mappings for error codes
func GetHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeInvalidCredentials, ErrCodeUnauthorized, ErrCodeTokenExpired, ErrCodeTokenInvalid, ErrCodeMissingAuth:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeInvalidRole:
		return http.StatusForbidden
	case ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeMissingField, ErrCodeInvalidFormat, ErrCodeInvalidQuantity:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists, ErrCodeConflict, ErrCodeInsufficientStock:
		return http.StatusConflict
	case ErrCodeInternalError, ErrCodeDatabaseError, ErrCodeServiceUnavailable:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// RespondWithAPIError is a convenience function that automatically determines the HTTP status
func RespondWithAPIError(c *gin.Context, err *APIError) {
	status := GetHTTPStatus(err.Code)
	RespondWithError(c, status, err)
}

// RespondWithAPIErrors is a convenience function for multiple errors
func RespondWithAPIErrors(c *gin.Context, errors []APIError) {
	if len(errors) == 0 {
		RespondWithAPIError(c, InternalError("Unknown error occurred"))
		return
	}
	
	status := GetHTTPStatus(errors[0].Code)
	RespondWithErrors(c, status, errors)
}
