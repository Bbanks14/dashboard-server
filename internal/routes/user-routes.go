// File: internal/routes/user_routes.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/dashboard-backend/internal/controllers"
	"github.com/yourusername/dashboard-backend/internal/middleware"
)

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(router *gin.Engine, controller *controllers.UserController, authMiddleware *middleware.AuthMiddleware) {
	users := router.Group("/api/users")
	{
		// Public routes
		users.GET("/public-profile/:id", controller.GetPublicProfile)

		// Protected routes
		authorized := users.Group("/")
		authorized.Use(authMiddleware.RequireAuth())
		{
			authorized.GET("/:id", controller.GetUser)
			authorized.POST("/", controller.CreateUser)
			authorized.PUT("/:id", controller.UpdateUser)
			authorized.DELETE("/:id", controller.DeleteUser)
			authorized.GET("/", controller.ListUsers)
		}
	}
}
