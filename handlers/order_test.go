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

func TestOrderHandler_CreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewOrderHandler()

	tests := []struct {
		name           string
		requestBody    models.CreateOrderRequest
		userID         int
		expectedStatus int
	}{
		{
			name: "valid order creation",
			requestBody: models.CreateOrderRequest{
				ProductID: 1,
				Quantity:  2,
			},
			userID:         2,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "insufficient stock",
			requestBody: models.CreateOrderRequest{
				ProductID: 1,
				Quantity:  20, // More than available stock (10)
			},
			userID:         2,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-existent product",
			requestBody: models.CreateOrderRequest{
				ProductID: 999,
				Quantity:  1,
			},
			userID:         2,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid quantity",
			requestBody: models.CreateOrderRequest{
				ProductID: 1,
				Quantity:  0,
			},
			userID:         2,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})
			r.POST("/orders", handler.CreateOrder)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.OrderResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.ProductID, response.Order.ProductID)
				assert.Equal(t, tt.requestBody.Quantity, response.Order.Quantity)
				assert.Equal(t, tt.userID, response.Order.UserID)
				assert.Equal(t, models.OrderStatusCompleted, response.Order.Status)
				assert.Greater(t, response.Order.Total, 0.0)
			}
		})
	}
}

func TestOrderHandler_GetUserOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewOrderHandler()

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", 2) // User with existing orders
		c.Next()
	})
	r.GET("/orders", handler.GetUserOrders)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var orders []models.OrderWithDetails
	err := json.Unmarshal(w.Body.Bytes(), &orders)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(orders), 1) // We have at least 1 test order
	
	// Check that all orders belong to the user
	for _, order := range orders {
		assert.Equal(t, 2, order.UserID)
		assert.NotEmpty(t, order.ProductName)
	}
}

func TestOrderHandler_GetOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewOrderHandler()

	tests := []struct {
		name           string
		orderID        string
		userID         int
		role           models.Role
		expectedStatus int
	}{
		{
			name:           "user gets own order",
			orderID:        "1",
			userID:         2,
			role:           models.RoleUser,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "admin gets any order",
			orderID:        "1",
			userID:         1,
			role:           models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user tries to get other user's order",
			orderID:        "1",
			userID:         3,
			role:           models.RoleUser,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid order ID",
			orderID:        "abc",
			userID:         2,
			role:           models.RoleUser,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent order",
			orderID:        "999",
			userID:         2,
			role:           models.RoleUser,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/orders/"+tt.orderID, nil)
			w := httptest.NewRecorder()

			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Set("role", tt.role)
				c.Next()
			})
			r.GET("/orders/:id", handler.GetOrder)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var order models.OrderWithDetails
				err := json.Unmarshal(w.Body.Bytes(), &order)
				assert.NoError(t, err)
				assert.NotEmpty(t, order.ProductName)
				assert.Greater(t, order.Total, 0.0)
			}
		})
	}
}

func TestOrderHandler_GetAllOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewOrderHandler()

	req := httptest.NewRequest("GET", "/admin/orders", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.GET("/admin/orders", handler.GetAllOrders)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var orders []models.OrderWithDetails
	err := json.Unmarshal(w.Body.Bytes(), &orders)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(orders), 1) // We have at least 1 test order
	
	// Check that orders have both product name and username
	for _, order := range orders {
		assert.NotEmpty(t, order.ProductName)
		assert.NotEmpty(t, order.Username)
	}
}

func TestOrderHandler_StockReduction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewOrderHandler()

	// Get initial stock
	initialProduct, err := testutil.GetTestProduct(t, 1)
	assert.NoError(t, err)
	initialStock := initialProduct["stock"].(int)

	// Create order
	orderReq := models.CreateOrderRequest{
		ProductID: 1,
		Quantity:  3,
	}
	jsonBody, _ := json.Marshal(orderReq)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", 2)
		c.Next()
	})
	r.POST("/orders", handler.CreateOrder)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Check that stock was reduced
	updatedProduct, err := testutil.GetTestProduct(t, 1)
	assert.NoError(t, err)
	updatedStock := updatedProduct["stock"].(int)
	
	assert.Equal(t, initialStock-3, updatedStock)
}
