package routes

import (
	"github.com/Bbanks14/dashboard-server/internal/middleware"
	"github.com/Bbanks14/dashboard-server/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures user-related routes
func SetupDashboardRoutes(router *gin.Engine, controller *controllers.DashboardController, authMiddleware *middleware.AuthMiddleware) {
	dashboard := router.Group("/api/dashboard")
	{
		// Public routes
		dashboard.GET("/users", controllers.DashboardController.GetUser())
		dashboard.GET("/stats", controllers.DashboardController.GetDashboardStats())
		dashboard.GET("/summary", controllers.DashboardController.GetDashboardSummary())

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
