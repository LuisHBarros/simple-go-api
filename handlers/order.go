package handlers

import (
	"database/sql"
	"net/http"
	"smarapp-api/database"
	"smarapp-api/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct{}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

// CreateOrder godoc
// @Summary Create a new order (Buy a product)
// @Description Create a new order to purchase a product, automatically reduces stock
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.CreateOrderRequest true "Order data"
// @Success 201 {object} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get product and check stock
	var product models.Product
	err = tx.QueryRow(
		"SELECT id, name, description, price, stock, created_by, created_at, updated_at FROM products WHERE id = ?",
		req.ProductID,
	).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.CreatedBy, &product.CreatedAt, &product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if enough stock
	if product.Stock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
		return
	}

	// Calculate total
	total := product.Price * float64(req.Quantity)

	// Create order
	result, err := tx.Exec(
		"INSERT INTO orders (user_id, product_id, quantity, price, total, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID, req.ProductID, req.Quantity, product.Price, total, models.OrderStatusPending, time.Now(), time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	orderID, _ := result.LastInsertId()

	// Update product stock
	_, err = tx.Exec(
		"UPDATE products SET stock = stock - ?, updated_at = ? WHERE id = ?",
		req.Quantity, time.Now(), req.ProductID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
		return
	}

	// Complete order (simulate payment success)
	_, err = tx.Exec(
		"UPDATE orders SET status = ?, updated_at = ? WHERE id = ?",
		models.OrderStatusCompleted, time.Now(), orderID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete order"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	order := models.Order{
		ID:        int(orderID),
		UserID:    userID.(int),
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     product.Price,
		Total:     total,
		Status:    models.OrderStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Update product stock for response
	product.Stock -= req.Quantity

	c.JSON(http.StatusCreated, models.OrderResponse{
		Order:   order,
		Product: product,
	})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := database.DB.Query(`
		SELECT o.id, o.user_id, o.product_id, o.quantity, o.price, o.total, o.status, o.created_at, o.updated_at,
		       p.name as product_name
		FROM orders o
		JOIN products p ON o.product_id = p.id
		WHERE o.user_id = ?
		ORDER BY o.created_at DESC
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()

	var orders []models.OrderWithDetails
	for rows.Next() {
		var order models.OrderWithDetails
		err := rows.Scan(
			&order.ID, &order.UserID, &order.ProductID, &order.Quantity,
			&order.Price, &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt,
			&order.ProductName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order"})
			return
		}
		orders = append(orders, order)
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT o.id, o.user_id, o.product_id, o.quantity, o.price, o.total, o.status, o.created_at, o.updated_at,
		       p.name as product_name, u.username
		FROM orders o
		JOIN products p ON o.product_id = p.id
		JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()

	var orders []models.OrderWithDetails
	for rows.Next() {
		var order models.OrderWithDetails
		err := rows.Scan(
			&order.ID, &order.UserID, &order.ProductID, &order.Quantity,
			&order.Price, &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt,
			&order.ProductName, &order.Username,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order"})
			return
		}
		orders = append(orders, order)
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	query := `
		SELECT o.id, o.user_id, o.product_id, o.quantity, o.price, o.total, o.status, o.created_at, o.updated_at,
		       p.name as product_name
		FROM orders o
		JOIN products p ON o.product_id = p.id
		WHERE o.id = ?`

	args := []interface{}{id}

	// If not admin, only show user's own orders
	if role != models.RoleAdmin {
		query += " AND o.user_id = ?"
		args = append(args, userID)
	}

	var order models.OrderWithDetails
	err = database.DB.QueryRow(query, args...).Scan(
		&order.ID, &order.UserID, &order.ProductID, &order.Quantity,
		&order.Price, &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt,
		&order.ProductName,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, order)
}
