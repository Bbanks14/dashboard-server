package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func GinAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenCookie, err := c.Cookie("jwt_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Parse token
		claims, err := ParseJWT(tokenCookie, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Add user ID to context
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}
