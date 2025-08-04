package errors

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleValidationError is a helper function to handle validation errors consistently
func HandleValidationError(c *gin.Context, err error) {
	validationErrors := FormatValidationErrors(err)
	RespondWithAPIErrors(c, validationErrors)
}

// Business validation helpers

// ValidateRole checks if a role is valid
func ValidateRole(role string) *APIError {
	validRoles := []string{"user", "admin"}
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}
	return InvalidInput("role", fmt.Sprintf("Role must be one of: %s", strings.Join(validRoles, ", ")))
}

// ValidatePositiveQuantity checks if quantity is positive
func ValidatePositiveQuantity(quantity int, field string) *APIError {
	if quantity <= 0 {
		return InvalidInput(field, fmt.Sprintf("%s must be greater than 0", field))
	}
	return nil
}

// ValidatePositivePrice checks if price is positive
func ValidatePositivePrice(price float64, field string) *APIError {
	if price <= 0 {
		return InvalidInput(field, fmt.Sprintf("%s must be greater than 0", field))
	}
	return nil
}

// ValidateNonNegativeStock checks if stock is non-negative
func ValidateNonNegativeStock(stock int, field string) *APIError {
	if stock < 0 {
		return InvalidInput(field, fmt.Sprintf("%s cannot be negative", field))
	}
	return nil
}

// ValidateStringLength checks string length constraints
func ValidateStringLength(value, field string, min, max int) *APIError {
	length := len(strings.TrimSpace(value))
	
	if min > 0 && length < min {
		return InvalidInput(field, fmt.Sprintf("%s must be at least %d characters long", field, min))
	}
	
	if max > 0 && length > max {
		return InvalidInput(field, fmt.Sprintf("%s must be at most %d characters long", field, max))
	}
	
	return nil
}

// ValidateEmail checks if email format is valid (basic check)
func ValidateEmail(email string) *APIError {
	email = strings.TrimSpace(email)
	if email == "" {
		return InvalidInput("email", "Email is required")
	}
	
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return InvalidInput("email", "Email must be a valid email address")
	}
	
	return nil
}

// ValidatePassword checks password requirements
func ValidatePassword(password string) *APIError {
	if len(password) < 6 {
		return InvalidInput("password", "Password must be at least 6 characters long")
	}
	
	// Add more password validation rules as needed
	// e.g., require uppercase, lowercase, numbers, special characters
	
	return nil
}

// ValidateID checks if an ID parameter is valid
func ValidateID(idStr, resource string) (int, *APIError) {
	if idStr == "" {
		return 0, InvalidInput("id", fmt.Sprintf("%s ID is required", resource))
	}
	
	// This would typically parse the ID, but since we're using string params
	// we'll assume it's already validated by the router
	// In a real implementation, you might want to parse and validate the integer
	
	return 0, nil // Placeholder - actual ID parsing would go here
}

// ValidateOrderRequest validates a complete order request
func ValidateOrderRequest(items []struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}) []APIError {
	var errors []APIError
	
	if len(items) == 0 {
		errors = append(errors, *InvalidInput("items", "At least one item is required"))
		return errors
	}
	
	for i, item := range items {
		field := fmt.Sprintf("items[%d]", i)
		
		if item.ProductID <= 0 {
			errors = append(errors, *InvalidInput(field+".product_id", "Product ID must be greater than 0"))
		}
		
		if item.Quantity <= 0 {
			errors = append(errors, *InvalidInput(field+".quantity", "Quantity must be greater than 0"))
		}
	}
	
	return errors
}

// Business logic validation errors

// InsufficientStock creates an error for insufficient stock
func InsufficientStock(productName string, requested, available int) *APIError {
	message := fmt.Sprintf("Insufficient stock for %s. Requested: %d, Available: %d", 
		productName, requested, available)
	return NewAPIError(ErrCodeInsufficientStock, message)
}

// ProductNotFound creates a product not found error
func ProductNotFound(productID interface{}) *APIError {
	return NotFound("Product", fmt.Sprintf("Product with ID %v does not exist", productID))
}

// UserNotFound creates a user not found error
func UserNotFound(identifier string) *APIError {
	return NotFound("User", fmt.Sprintf("User %s does not exist", identifier))
}

// OrderNotFound creates an order not found error
func OrderNotFound(orderID interface{}) *APIError {
	return NotFound("Order", fmt.Sprintf("Order with ID %v does not exist", orderID))
}

// UserAlreadyExists creates a user already exists error
func UserAlreadyExists(field, value string) *APIError {
	message := fmt.Sprintf("User with this %s already exists", field)
	details := fmt.Sprintf("A user with %s '%s' is already registered", field, value)
	return NewAPIError(ErrCodeAlreadyExists, message, details)
}

// InvalidCredentials creates an invalid credentials error
func InvalidCredentials() *APIError {
	return NewAPIError(ErrCodeInvalidCredentials, 
		"Invalid email or password", 
		"Please check your credentials and try again")
}

// TokenExpired creates a token expired error
func TokenExpired() *APIError {
	return NewAPIError(ErrCodeTokenExpired, 
		"Your session has expired", 
		"Please log in again to continue")
}

// InvalidToken creates an invalid token error
func InvalidToken() *APIError {
	return NewAPIError(ErrCodeTokenInvalid, 
		"Invalid authentication token", 
		"Please log in again to continue")
}

// MissingAuthHeader creates a missing authorization header error
func MissingAuthHeader() *APIError {
	return NewAPIError(ErrCodeMissingAuth, 
		"Authorization header is required", 
		"Please include a valid Bearer token in the Authorization header")
}

// InvalidAuthFormat creates an invalid authorization format error
func InvalidAuthFormat() *APIError {
	return NewAPIError(ErrCodeMissingAuth, 
		"Invalid authorization format", 
		"Authorization header must be in format: Bearer <token>")
}

// AdminRequired creates an admin access required error
func AdminRequired() *APIError {
	return NewAPIError(ErrCodeForbidden, 
		"Administrator access required", 
		"This operation requires administrator privileges")
}

// UserRoleNotFound creates a user role not found error
func UserRoleNotFound() *APIError {
	return NewAPIError(ErrCodeUnauthorized, 
		"User role information not found", 
		"Please log in again to continue")
}

