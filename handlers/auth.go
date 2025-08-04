package handlers

import (
	"database/sql"
	"net/http"
	"smarapp-api/database"
	"smarapp-api/errors"
	"smarapp-api/middleware"
	"smarapp-api/models"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	JWTSecret string
}

func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{
		JWTSecret: jwtSecret,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username, email, password and optional role
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use the new error handling system for validation errors
		validationErrors := errors.FormatValidationErrors(err)
		errors.RespondWithAPIErrors(c, validationErrors)
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = models.RoleUser
	}

	// Validate role using the new error system
	if req.Role != models.RoleUser && req.Role != models.RoleAdmin {
		apiErr := errors.InvalidInput("role", "Role must be either 'user' or 'admin'")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		apiErr := errors.InternalError("Failed to process password", "Password hashing failed")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	// Check if user already exists
	var existingID int
	err = database.DB.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", req.Email, req.Username).Scan(&existingID)
	if err != sql.ErrNoRows {
		// Determine which field conflicts
		var emailExists, usernameExists bool
		database.DB.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existingID)
		if existingID > 0 {
			emailExists = true
		}
		database.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
		if existingID > 0 {
			usernameExists = true
		}

		if emailExists && usernameExists {
			apiErr := errors.AlreadyExists("User", "Both email and username are already registered")
			errors.RespondWithAPIError(c, apiErr)
		} else if emailExists {
			apiErr := errors.UserAlreadyExists("email", req.Email)
			errors.RespondWithAPIError(c, apiErr)
		} else {
			apiErr := errors.UserAlreadyExists("username", req.Username)
			errors.RespondWithAPIError(c, apiErr)
		}
		return
	}

	// Insert user
	result, err := database.DB.Exec(
		"INSERT INTO users (username, email, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		req.Username, req.Email, string(hashedPassword), req.Role, time.Now(), time.Now(),
	)
	if err != nil {
		apiErr := errors.DatabaseError("user creation", "Failed to save user to database")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	userID, _ := result.LastInsertId()

	user := models.User{
		ID:        int(userID),
		Username:  req.Username,
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate token
	token, err := middleware.GenerateToken(user, h.JWTSecret)
	if err != nil {
		apiErr := errors.InternalError("Failed to generate authentication token", "Token generation failed")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := errors.FormatValidationErrors(err)
		errors.RespondWithAPIErrors(c, validationErrors)
		return
	}

	var user models.User
	var hashedPassword string
	err := database.DB.QueryRow(
		"SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		apiErr := errors.InvalidCredentials()
		errors.RespondWithAPIError(c, apiErr)
		return
	}
	if err != nil {
		apiErr := errors.DatabaseError("user lookup", "Failed to retrieve user information")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		apiErr := errors.InvalidCredentials()
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(user, h.JWTSecret)
	if err != nil {
		apiErr := errors.InternalError("Failed to generate authentication token", "Token generation failed")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apiErr := errors.UserRoleNotFound()
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, role, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		apiErr := errors.UserNotFound("with the provided ID")
		errors.RespondWithAPIError(c, apiErr)
		return
	}
	if err != nil {
		apiErr := errors.DatabaseError("profile retrieval", "Failed to retrieve user profile")
		errors.RespondWithAPIError(c, apiErr)
		return
	}

	c.JSON(http.StatusOK, user)
}
