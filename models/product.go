package models

import (
	"time"
)

type Product struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	CreatedBy   int       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"required,min=1,max=500"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description string  `json:"description,omitempty" binding:"omitempty,min=1,max=500"`
	Price       float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
	Stock       int     `json:"stock,omitempty" binding:"omitempty,gte=0"`
}
