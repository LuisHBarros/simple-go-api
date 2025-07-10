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

func TestProductHandler_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	handler := NewProductHandler()

	tests := []struct {
		name           string
		requestBody    models.CreateProductRequest
		userID         int
		expectedStatus int
	}{
		{
			name: "valid product creation",
			requestBody: models.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
			},
			userID:         1,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid product - empty name",
			requestBody: models.CreateProductRequest{
				Name:        "",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
			},
			userID:         1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid product - zero price",
			requestBody: models.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       0,
				Stock:       10,
			},
			userID:         1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid product - negative stock",
			requestBody: models.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       -1,
			},
			userID:         1,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})
			r.POST("/products", handler.CreateProduct)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var product models.Product
				err := json.Unmarshal(w.Body.Bytes(), &product)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.Name, product.Name)
				assert.Equal(t, tt.requestBody.Description, product.Description)
				assert.Equal(t, tt.requestBody.Price, product.Price)
				assert.Equal(t, tt.requestBody.Stock, product.Stock)
				assert.Equal(t, tt.userID, product.CreatedBy)
			}
		})
	}
}

func TestProductHandler_GetProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewProductHandler()

	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.GET("/products", handler.GetProducts)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	err := json.Unmarshal(w.Body.Bytes(), &products)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 2) // We have 2 test products
}

func TestProductHandler_GetProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewProductHandler()

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "valid product ID",
			productID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid product ID",
			productID:      "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent product",
			productID:      "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()

			r := gin.New()
			r.GET("/products/:id", handler.GetProduct)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var product models.Product
				err := json.Unmarshal(w.Body.Bytes(), &product)
				assert.NoError(t, err)
				assert.NotEmpty(t, product.Name)
				assert.Greater(t, product.Price, 0.0)
			}
		})
	}
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewProductHandler()

	tests := []struct {
		name           string
		productID      string
		requestBody    models.UpdateProductRequest
		expectedStatus int
	}{
		{
			name:      "valid update",
			productID: "1",
			requestBody: models.UpdateProductRequest{
				Name:  "Updated Product",
				Price: 149.99,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "partial update",
			productID: "1",
			requestBody: models.UpdateProductRequest{
				Stock: 20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "invalid product ID",
			productID: "abc",
			requestBody: models.UpdateProductRequest{
				Name: "Updated Product",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "non-existent product",
			productID: "999",
			requestBody: models.UpdateProductRequest{
				Name: "Updated Product",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/products/"+tt.productID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r := gin.New()
			r.PUT("/products/:id", handler.UpdateProduct)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var product models.Product
				err := json.Unmarshal(w.Body.Bytes(), &product)
				assert.NoError(t, err)
				
				if tt.requestBody.Name != "" {
					assert.Equal(t, tt.requestBody.Name, product.Name)
				}
				if tt.requestBody.Price > 0 {
					assert.Equal(t, tt.requestBody.Price, product.Price)
				}
				if tt.requestBody.Stock >= 0 {
					assert.Equal(t, tt.requestBody.Stock, product.Stock)
				}
			}
		})
	}
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cleanup := testutil.SetupTestDBWithData(t)
	defer cleanup()

	handler := NewProductHandler()

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "valid deletion",
			productID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid product ID",
			productID:      "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent product",
			productID:      "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()

			r := gin.New()
			r.DELETE("/products/:id", handler.DeleteProduct)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), "deleted successfully")
			}
		})
	}
}
