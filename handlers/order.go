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

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		apiErr := errors.DatabaseError("transaction start", "Failed to begin database transaction")
		errors.RespondWithAPIError(c, apiErr)
		return
	}
	defer tx.Rollback()

	var orderTotal float64
	var products []models.Product
	var orderItems []models.OrderItem

	// Validate all products and calculate total
	for _, item := range req.Items {
		// Validate quantity
		if item.Quantity <= 0 {
			apiErr := errors.InvalidInput("quantity", "Quantity must be greater than 0")
			errors.RespondWithAPIError(c, apiErr)
			return
		}

		var product models.Product
		err = tx.QueryRow(
			"SELECT id, name, description, price, stock, created_by, created_at, updated_at FROM products WHERE id = ?",
			item.ProductID,
		).Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.CreatedBy, &product.CreatedAt, &product.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			apiErr := errors.ProductNotFound(item.ProductID)
			errors.RespondWithAPIError(c, apiErr)
			return
		}
		if err != nil {
			apiErr := errors.DatabaseError("product lookup", "Failed to retrieve product information")
			errors.RespondWithAPIError(c, apiErr)
			return
		}

		if product.Stock < item.Quantity {
			apiErr := errors.InsufficientStock(product.Name, item.Quantity, product.Stock)
			errors.RespondWithAPIError(c, apiErr)
			return
		}

		itemTotal := product.Price * float64(item.Quantity)
		orderTotal += itemTotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Total:     itemTotal,
		})

		product.Stock -= item.Quantity
		products = append(products, product)
	}

	// Create order
	result, err := tx.Exec(
		"INSERT INTO orders (user_id, total, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		userID, orderTotal, models.OrderStatusPending, time.Now(), time.Now(),
	)
	if err != nil {
		apiErr := errors.DatabaseError("order creation", "Failed to create order")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	orderID, _ := result.LastInsertId()

	// Create order items and update stock
	for i, item := range orderItems {
		_, err = tx.Exec(
			"INSERT INTO order_items (order_id, product_id, quantity, price, total) VALUES (?, ?, ?, ?, ?)",
			orderID, item.ProductID, item.Quantity, item.Price, item.Total,
		)
		if err != nil {
			apiErr := errors.DatabaseError("order item creation", "Failed to create order item")
			errors.RespondWithAPIError(c, apiErr)
			return
		}

		_, err = tx.Exec(
			"UPDATE products SET stock = stock - ?, updated_at = ? WHERE id = ?",
			item.Quantity, time.Now(), item.ProductID,
		)
		if err != nil {
			apiErr := errors.DatabaseError("stock update", "Failed to update product stock")
			errors.RespondWithAPIError(c, apiErr)
			return
		}

		orderItems[i].OrderID = int(orderID)
	}

	// Complete order
	_, err = tx.Exec(
		"UPDATE orders SET status = ?, updated_at = ? WHERE id = ?",
		models.OrderStatusCompleted, time.Now(), orderID,
	)
	if err != nil {
		apiErr := errors.DatabaseError("order completion", "Failed to complete order")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	if err = tx.Commit(); err != nil {
		apiErr := errors.DatabaseError("transaction commit", "Failed to commit order transaction")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	order := models.Order{
		ID:        int(orderID),
		UserID:    userID.(int),
		Total:     orderTotal,
		Status:    models.OrderStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Items:     orderItems,
	}

	c.JSON(http.StatusCreated, models.OrderResponse{
		Order:    order,
		Products: products,
	})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := database.DB.Query(`
		SELECT o.id, o.user_id, o.total, o.status, o.created_at, o.updated_at, u.username
		FROM orders o
		JOIN users u ON o.user_id = u.id
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
		var orderID, orderUserID int
		var total float64
		var status models.OrderStatus
		var createdAt, updatedAt time.Time
		var username string

		err := rows.Scan(
			&orderID, &orderUserID, &total, &status, &createdAt, &updatedAt, &username,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order"})
			return
		}

		// Get order items with product names
		itemRows, err := database.DB.Query(`
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.total, p.name as product_name
			FROM order_items oi
			JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = ?
		`, orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
			return
		}

		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.Total, &item.ProductName)
			if err != nil {
				itemRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order item"})
				return
			}
			items = append(items, item)
		}
		itemRows.Close()

		orderWithDetails := models.OrderWithDetails{
			Order: models.Order{
				ID:        orderID,
				UserID:    orderUserID,
				Total:     total,
				Status:    status,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Items:     items,
			},
			Username: username,
		}

		orders = append(orders, orderWithDetails)
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT o.id, o.user_id, o.total, o.status, o.created_at, o.updated_at, u.username
		FROM orders o
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
		var orderID, orderUserID int
		var total float64
		var status models.OrderStatus
		var createdAt, updatedAt time.Time
		var username string

		err := rows.Scan(
			&orderID, &orderUserID, &total, &status, &createdAt, &updatedAt, &username,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order"})
			return
		}

		// Get order items with product names
		itemRows, err := database.DB.Query(`
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.total, p.name as product_name
			FROM order_items oi
			JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = ?
		`, orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
			return
		}

		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.Total, &item.ProductName)
			if err != nil {
				itemRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order item"})
				return
			}
			items = append(items, item)
		}
		itemRows.Close()

		orderWithDetails := models.OrderWithDetails{
			Order: models.Order{
				ID:        orderID,
				UserID:    orderUserID,
				Total:     total,
				Status:    status,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Items:     items,
			},
			Username: username,
		}

		orders = append(orders, orderWithDetails)
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
		SELECT o.id, o.user_id, o.total, o.status, o.created_at, o.updated_at, u.username
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = ?` + func() string {
			if role != models.RoleAdmin {
				return " AND o.user_id = ?"
			}
			return ""
		}() + `
	`

	args := []interface{}{id}
	if role != models.RoleAdmin {
		args = append(args, userID)
	}

	var orderID, orderUserID int
	var total float64
	var status models.OrderStatus
	var createdAt, updatedAt time.Time
	var username string

	err = database.DB.QueryRow(query, args...).Scan(
		&orderID, &orderUserID, &total, &status, &createdAt, &updatedAt, &username,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Get order items with product names
	itemRows, err := database.DB.Query(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.total, p.name as product_name
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = ?
	`, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
		return
	}
	defer itemRows.Close()

	var items []models.OrderItem
	for itemRows.Next() {
		var item models.OrderItem
		err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.Total, &item.ProductName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order item"})
			return
		}
		items = append(items, item)
	}

	orderWithDetails := models.OrderWithDetails{
		Order: models.Order{
			ID:        orderID,
			UserID:    orderUserID,
			Total:     total,
			Status:    status,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Items:     items,
		},
		Username: username,
	}

	c.JSON(http.StatusOK, orderWithDetails)
}
