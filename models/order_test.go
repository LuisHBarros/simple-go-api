package models

import (
	"testing"

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
		ProductID: 3,
		Quantity:  5,
		Price:     99.99,
		Total:     499.95,
		Status:    OrderStatusCompleted,
	}

	assert.Equal(t, 1, order.ID)
	assert.Equal(t, 2, order.UserID)
	assert.Equal(t, 3, order.ProductID)
	assert.Equal(t, 5, order.Quantity)
	assert.Equal(t, 99.99, order.Price)
	assert.Equal(t, 499.95, order.Total)
	assert.Equal(t, OrderStatusCompleted, order.Status)
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
				ProductID: 1,
				Quantity:  2,
			},
			valid: true,
		},
		{
			name: "zero product ID",
			request: CreateOrderRequest{
				ProductID: 0,
				Quantity:  2,
			},
			valid: false,
		},
		{
			name: "zero quantity",
			request: CreateOrderRequest{
				ProductID: 1,
				Quantity:  0,
			},
			valid: false,
		},
		{
			name: "negative quantity",
			request: CreateOrderRequest{
				ProductID: 1,
				Quantity:  -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.Greater(t, tt.request.ProductID, 0)
				assert.Greater(t, tt.request.Quantity, 0)
			} else {
				assert.True(t, tt.request.ProductID <= 0 || tt.request.Quantity <= 0)
			}
		})
	}
}

func TestOrderResponse_Structure(t *testing.T) {
	order := Order{
		ID:        1,
		UserID:    2,
		ProductID: 3,
		Quantity:  2,
		Price:     99.99,
		Total:     199.98,
		Status:    OrderStatusCompleted,
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
		Order:   order,
		Product: product,
	}

	assert.Equal(t, order, response.Order)
	assert.Equal(t, product, response.Product)
	assert.Equal(t, order.ProductID, product.ID)
	assert.Equal(t, order.Price, product.Price)
}

func TestOrderWithDetails_Structure(t *testing.T) {
	orderWithDetails := OrderWithDetails{
		Order: Order{
			ID:        1,
			UserID:    2,
			ProductID: 3,
			Quantity:  2,
			Price:     99.99,
			Total:     199.98,
			Status:    OrderStatusCompleted,
		},
		ProductName: "Test Product",
		Username:    "testuser",
	}

	assert.Equal(t, 1, orderWithDetails.Order.ID)
	assert.Equal(t, "Test Product", orderWithDetails.ProductName)
	assert.Equal(t, "testuser", orderWithDetails.Username)
}
