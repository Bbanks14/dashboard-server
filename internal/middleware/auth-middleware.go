// File: internal/middleware/auth_middleware.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/dashboard-backend/internal/services"
	"github.com/yourusername/dashboard-backend/internal/util/config"
	"github.com/yourusername/dashboard-backend/pkg/helpers"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	authService *services.AuthService
	config      *config.Config
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *services.AuthService, config *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		config:      config,
	}
}

// RequireAuth ensures a valid JWT token is present
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			helpers.ErrorResponse(ctx, http.StatusUnauthorized, "Authorization header required")
			ctx.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			helpers.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid authorization format")
			ctx.Abort()
			return
		}

		token := parts[1]
		user, err := m.authService.ValidateToken(token)
		if err != nil {
			helpers.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid or expired token")
			ctx.Abort()
			return
		}

		// Set user in context for use in handlers
		ctx.Set("user", user)
		ctx.Next()
	}
}
