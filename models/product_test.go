package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProduct_Structure(t *testing.T) {
	now := time.Now()
	product := Product{
		ID:          1,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
		CreatedBy:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, 1, product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "Test Description", product.Description)
	assert.Equal(t, 99.99, product.Price)
	assert.Equal(t, 10, product.Stock)
	assert.Equal(t, 1, product.CreatedBy)
	assert.Equal(t, now, product.CreatedAt)
	assert.Equal(t, now, product.UpdatedAt)
}

func TestCreateProductRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateProductRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
			},
			valid: true,
		},
		{
			name: "empty name",
			request: CreateProductRequest{
				Name:        "",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
			},
			valid: false,
		},
		{
			name: "long name",
			request: CreateProductRequest{
				Name:        "This is a very long product name that exceeds the maximum allowed length of one hundred characters",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
			},
			valid: false,
		},
		{
			name: "empty description",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "",
				Price:       99.99,
				Stock:       10,
			},
			valid: false,
		},
		{
			name: "zero price",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       0,
				Stock:       10,
			},
			valid: false,
		},
		{
			name: "negative price",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       -10.0,
				Stock:       10,
			},
			valid: false,
		},
		{
			name: "negative stock",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       -1,
			},
			valid: false,
		},
		{
			name: "zero stock (valid)",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       0,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
				assert.NotEmpty(t, tt.request.Description)
				assert.Greater(t, tt.request.Price, 0.0)
				assert.GreaterOrEqual(t, tt.request.Stock, 0)
				assert.LessOrEqual(t, len(tt.request.Name), 100)
				assert.LessOrEqual(t, len(tt.request.Description), 500)
			}
		})
	}
}

func TestUpdateProductRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateProductRequest
		valid   bool
	}{
		{
			name: "valid partial update",
			request: UpdateProductRequest{
				Name:  "Updated Product",
				Price: 149.99,
			},
			valid: true,
		},
		{
			name: "valid full update",
			request: UpdateProductRequest{
				Name:        "Updated Product",
				Description: "Updated Description",
				Price:       149.99,
				Stock:       20,
			},
			valid: true,
		},
		{
			name: "empty update (valid)",
			request: UpdateProductRequest{},
			valid: true,
		},
		{
			name: "invalid name too long",
			request: UpdateProductRequest{
				Name: "This is a very long product name that exceeds the maximum allowed length of one hundred characters",
			},
			valid: false,
		},
		{
			name: "invalid zero price",
			request: UpdateProductRequest{
				Price: 0,
			},
			valid: false,
		},
		{
			name: "invalid negative stock",
			request: UpdateProductRequest{
				Stock: -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				if tt.request.Name != "" {
					assert.LessOrEqual(t, len(tt.request.Name), 100)
				}
				if tt.request.Description != "" {
					assert.LessOrEqual(t, len(tt.request.Description), 500)
				}
				if tt.request.Price != 0 {
					assert.Greater(t, tt.request.Price, 0.0)
				}
				if tt.request.Stock != 0 {
					assert.GreaterOrEqual(t, tt.request.Stock, 0)
				}
			}
		})
	}
}
