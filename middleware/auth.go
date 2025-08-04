package middleware

import (
	"errors"
	apierrors "smarapp-api/errors"
	"smarapp-api/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int         `json:"user_id"`
	Username string      `json:"username"`
	Email    string      `json:"email"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User, jwtSecret string) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apiErr := apierrors.MissingAuthHeader()
			apierrors.RespondWithAPIError(c, apiErr)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			apiErr := apierrors.InvalidAuthFormat()
			apierrors.RespondWithAPIError(c, apiErr)
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			// Handle different types of token errors
			if err != nil {
				switch {
				case errors.Is(err, jwt.ErrTokenExpired):
					apiErr := apierrors.TokenExpired()
					apierrors.RespondWithAPIError(c, apiErr)
				case errors.Is(err, jwt.ErrTokenMalformed):
					apiErr := apierrors.InvalidToken()
					apierrors.RespondWithAPIError(c, apiErr)
				case errors.Is(err, jwt.ErrTokenNotValidYet):
					apiErr := apierrors.InvalidToken()
					apierrors.RespondWithAPIError(c, apiErr)
				default:
					apiErr := apierrors.InvalidToken()
					apierrors.RespondWithAPIError(c, apiErr)
				}
			} else {
				apiErr := apierrors.InvalidToken()
				apierrors.RespondWithAPIError(c, apiErr)
			}
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			apiErr := apierrors.UserRoleNotFound()
			apierrors.RespondWithAPIError(c, apiErr)
			c.Abort()
			return
		}

		if role != models.RoleAdmin {
			apiErr := apierrors.AdminRequired()
			apierrors.RespondWithAPIError(c, apiErr)
			c.Abort()
			return
		}

		c.Next()
	}
}
