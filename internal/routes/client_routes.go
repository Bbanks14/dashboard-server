package routes

import (
	"github.com/Bbanks14/dashboard-server/internal/controllers"
	"github.com/Bbanks14/dashboard-server/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupClientRoutes configures client-related routes
func SetupClientRoutes(router *gin.Engine, controller *controllers.ClientController, authMiddleware *middleware.AuthMiddleware) {
	clients := router.Group("/api/clients")
	{
		// Public routes
		clients.GET("/", controller.GetClients)

		// Protected routes
		authorized := clients.Group("/")
		authorized.Use(authMiddleware.RequireAuth())
		{
			authorized.GET("/products", controller.GetProducts)
			authorized.GET("/customers", controller.GetCustomers)
			authorized.GET("/transactions", controller.GetTransactions)
			authorized.GET("/geography", controller.GetGeography)
		}
	}
}
