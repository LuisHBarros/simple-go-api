package handlers

import (
	"database/sql"
	"net/http"
	"smarapp-api/database"
	"smarapp-api/errors"
	"smarapp-api/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// CreateProduct godoc
// @Summary Create a new product (Admin only)
// @Description Create a new product with name, description, price and stock
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.CreateProductRequest true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := errors.FormatValidationErrors(err)
		errors.RespondWithAPIErrors(c, validationErrors)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		apiErr := errors.UserRoleNotFound()
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO products (name, description, price, stock, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.Description, req.Price, req.Stock, userID, time.Now(), time.Now(),
	)
	if err != nil {
		apiErr := errors.DatabaseError("product creation", "Failed to save product to database")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	productID, _ := result.LastInsertId()

	product := models.Product{
		ID:          int(productID),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedBy:   userID.(int),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	c.JSON(http.StatusCreated, product)
}

// GetProducts godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags Products
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {object} map[string]string
// @Router /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, name, description, price, stock, created_by, created_at, updated_at FROM products ORDER BY created_at DESC",
	)
	if err != nil {
		apiErr := errors.DatabaseError("product retrieval", "Failed to retrieve products from database")
		errors.RespondWithAPIError(c, apiErr)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.CreatedBy, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			apiErr := errors.DatabaseError("product scanning", "Failed to process product data")
			errors.RespondWithAPIError(c, apiErr)
			return
		}
		products = append(products, product)
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		apiErr := errors.InvalidInput("id", "Product ID must be a valid number")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	var product models.Product
	err = database.DB.QueryRow(
		"SELECT id, name, description, price, stock, created_by, created_at, updated_at FROM products WHERE id = ?",
		id,
	).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.CreatedBy, &product.CreatedAt, &product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		apiErr := errors.ProductNotFound(id)
		errors.RespondWithAPIError(c, apiErr)
		return
	}
	if err != nil {
		apiErr := errors.DatabaseError("product retrieval", "Failed to retrieve product information")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		apiErr := errors.InvalidInput("id", "Product ID must be a valid number")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := errors.FormatValidationErrors(err)
		errors.RespondWithAPIErrors(c, validationErrors)
		return
	}

	// Check if product exists
	var existingProduct models.Product
	err = database.DB.QueryRow("SELECT id FROM products WHERE id = ?", id).Scan(&existingProduct.ID)
	if err == sql.ErrNoRows {
		apiErr := errors.ProductNotFound(id)
		errors.RespondWithAPIError(c, apiErr)
		return
	}
	if err != nil {
		apiErr := errors.DatabaseError("product lookup", "Failed to verify product existence")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	// Build dynamic update query
	query := "UPDATE products SET updated_at = ?"
	args := []interface{}{time.Now()}

	if req.Name != "" {
		query += ", name = ?"
		args = append(args, req.Name)
	}
	if req.Description != "" {
		query += ", description = ?"
		args = append(args, req.Description)
	}
	if req.Price > 0 {
		query += ", price = ?"
		args = append(args, req.Price)
	}
	if req.Stock >= 0 {
		query += ", stock = ?"
		args = append(args, req.Stock)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		apiErr := errors.DatabaseError("product update", "Failed to update product information")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	// Fetch updated product
	h.GetProduct(c)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		apiErr := errors.InvalidInput("id", "Product ID must be a valid number")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	result, err := database.DB.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		apiErr := errors.DatabaseError("product deletion", "Failed to delete product from database")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		apiErr := errors.ProductNotFound(id)
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
