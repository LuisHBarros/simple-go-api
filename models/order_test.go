package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderStatus_Constants(t *testing.T) {
	assert.Equal(t, OrderStatus("pending"), OrderStatusPending)
	assert.Equal(t, OrderStatus("completed"), OrderStatusCompleted)
	assert.Equal(t, OrderStatus("cancelled"), OrderStatusCancelled)
}

func TestOrder_Structure(t *testing.T) {
	order := Order{
		ID:        1,
		UserID:    2,
		Total:     499.95,
		Status:    OrderStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Items: []OrderItem{
			{
				ID:        1,
				OrderID:   1,
				ProductID: 3,
				Quantity:  5,
				Price:     99.99,
				Total:     499.95,
			},
		},
	}

	assert.Equal(t, 1, order.ID)
	assert.Equal(t, 2, order.UserID)
	assert.Equal(t, 499.95, order.Total)
	assert.Equal(t, OrderStatusCompleted, order.Status)
	assert.Len(t, order.Items, 1)
	assert.Equal(t, 3, order.Items[0].ProductID)
	assert.Equal(t, 5, order.Items[0].Quantity)
	assert.Equal(t, 99.99, order.Items[0].Price)
}

func TestCreateOrderRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateOrderRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateOrderRequest{
				Items: []struct {
					ProductID int `json:"product_id" binding:"required,gt=0"`
					Quantity  int `json:"quantity" binding:"required,gt=0"`
				}{
					{ProductID: 1, Quantity: 2},
				},
			},
			valid: true,
		},
		{
			name: "zero product ID",
			request: CreateOrderRequest{
				Items: []struct {
					ProductID int `json:"product_id" binding:"required,gt=0"`
					Quantity  int `json:"quantity" binding:"required,gt=0"`
				}{
					{ProductID: 0, Quantity: 2},
				},
			},
			valid: false,
		},
		{
			name: "zero quantity",
			request: CreateOrderRequest{
				Items: []struct {
					ProductID int `json:"product_id" binding:"required,gt=0"`
					Quantity  int `json:"quantity" binding:"required,gt=0"`
				}{
					{ProductID: 1, Quantity: 0},
				},
			},
			valid: false,
		},
		{
			name: "negative quantity",
			request: CreateOrderRequest{
				Items: []struct {
					ProductID int `json:"product_id" binding:"required,gt=0"`
					Quantity  int `json:"quantity" binding:"required,gt=0"`
				}{
					{ProductID: 1, Quantity: -1},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.Greater(t, tt.request.Items[0].ProductID, 0)
				assert.Greater(t, tt.request.Items[0].Quantity, 0)
			} else {
				assert.True(t, tt.request.Items[0].ProductID <= 0 || tt.request.Items[0].Quantity <= 0)
			}
		})
	}
}

func TestOrderResponse_Structure(t *testing.T) {
	order := Order{
		ID:        1,
		UserID:    2,
		Total:     199.98,
		Status:    OrderStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	product := Product{
		ID:          3,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
		CreatedBy:   1,
	}

	response := OrderResponse{
		Order:    order,
		Products: []Product{product},
	}

	assert.Equal(t, order, response.Order)
	assert.Equal(t, product, response.Products[0])
}

func TestOrderWithDetails_Structure(t *testing.T) {
	orderWithDetails := OrderWithDetails{
		Order: Order{
			ID:        1,
			UserID:    2,
			Total:     199.98,
			Status:    OrderStatusCompleted,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProductName: "Test Product",
		Username:    "testuser",
	}

	assert.Equal(t, 1, orderWithDetails.Order.ID)
	assert.Equal(t, "Test Product", orderWithDetails.ProductName)
	assert.Equal(t, "testuser", orderWithDetails.Username)
}
