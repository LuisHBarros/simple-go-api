package models

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID        int     `json:"id" db:"id"`
	OrderID   int     `json:"order_id" db:"order_id"`
	ProductID int     `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"` // Price at time of purchase
	Total     float64 `json:"total" db:"total"`
}

type Order struct {
	ID        int         `json:"id" db:"id"`
	UserID    int         `json:"user_id" db:"user_id"`
	Total     float64     `json:"total" db:"total"`
	Status    OrderStatus `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
	Items     []OrderItem `json:"items,omitempty"`
}

type OrderWithDetails struct {
	Order
	ProductName string `json:"product_name"`
	Username    string `json:"username"`
}

type CreateOrderRequest struct {
	Items []struct {
		ProductID int `json:"product_id" binding:"required,gt=0"`
		Quantity  int `json:"quantity" binding:"required,gt=0"`
	} `json:"items" binding:"required,min=1"`
}

type OrderResponse struct {
	Order    Order     `json:"order"`
	Products []Product `json:"products"`
}
