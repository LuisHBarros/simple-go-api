package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"smarapp-api/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates a temporary test database and returns a cleanup function
func SetupTestDB(t *testing.T) func() {
	// Create a temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Initialize the test database
	err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Return cleanup function
	return func() {
		database.CloseDB()
		os.Remove(dbPath)
	}
}

// SetupTestDBWithData creates a test database and populates it with test data
func SetupTestDBWithData(t *testing.T) func() {
	cleanup := SetupTestDB(t)

	// Insert test users
	_, err := database.DB.Exec(`
		INSERT INTO users (id, username, email, password, role, created_at, updated_at) 
		VALUES 
		(1, 'admin', 'admin@test.com', '$2a$10$test.hash.admin', 'admin', datetime('now'), datetime('now')),
		(2, 'user', 'user@test.com', '$2a$10$test.hash.user', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Insert test products
	_, err = database.DB.Exec(`
		INSERT INTO products (id, name, description, price, stock, created_by, created_at, updated_at)
		VALUES 
		(1, 'Test Product 1', 'Test Description 1', 99.99, 10, 1, datetime('now'), datetime('now')),
		(2, 'Test Product 2', 'Test Description 2', 149.99, 5, 1, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test products: %v", err)
	}

	// Insert test orders
	_, err = database.DB.Exec(`
		INSERT INTO orders (id, user_id, product_id, quantity, price, total, status, created_at, updated_at)
		VALUES 
		(1, 2, 1, 2, 99.99, 199.98, 'completed', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test orders: %v", err)
	}

	// Insert test chat messages
	_, err = database.DB.Exec(`
		INSERT INTO chat_messages (id, user_id, username, message, created_at)
		VALUES 
		(1, 2, 'user', 'Hello, world!', datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test chat messages: %v", err)
	}

	return cleanup
}

// GetTestUser returns a test user by ID
func GetTestUser(t *testing.T, userID int) (map[string]interface{}, error) {
	var id int
	var username, email, role string
	var createdAt, updatedAt string

	err := database.DB.QueryRow(
		"SELECT id, username, email, role, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&id, &username, &email, &role, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         id,
		"username":   username,
		"email":      email,
		"role":       role,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}, nil
}

// GetTestProduct returns a test product by ID
func GetTestProduct(t *testing.T, productID int) (map[string]interface{}, error) {
	var id, stock, createdBy int
	var name, description string
	var price float64
	var createdAt, updatedAt string

	err := database.DB.QueryRow(
		"SELECT id, name, description, price, stock, created_by, created_at, updated_at FROM products WHERE id = ?",
		productID,
	).Scan(&id, &name, &description, &price, &stock, &createdBy, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"price":       price,
		"stock":       stock,
		"created_by":  createdBy,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
	}, nil
}

// CountRows returns the number of rows in a table
func CountRows(t *testing.T, tableName string) int {
	var count int
	err := database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows in %s: %v", tableName, err)
	}
	return count
}

// ClearTable removes all rows from a table
func ClearTable(t *testing.T, tableName string) {
	_, err := database.DB.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	if err != nil {
		t.Fatalf("Failed to clear table %s: %v", tableName, err)
	}
}
