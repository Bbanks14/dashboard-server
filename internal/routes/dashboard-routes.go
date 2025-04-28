// File: internal/routes/dashboard_routes.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/dashboard-backend/internal/controllers"
	"github.com/yourusername/dashboard-backend/internal/middleware"
)

// SetupDashboardRoutes configures dashboard-related routes
func SetupDashboardRoutes(router *gin.Engine, controller *controllers.DashboardController, authMiddleware *middleware.AuthMiddleware) {
	dashboard := router.Group("/api/dashboard")
	dashboard.Use(authMiddleware.RequireAuth())
	{
		dashboard.GET("/stats", controller.GetDashboardStats)
		dashboard.GET("/summary", controller.GetDashboardSummary)
	}
}
