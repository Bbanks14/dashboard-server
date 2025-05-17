package api

import (
	"github.com/Bbanks14/dashboard-server/internal/controllers"

	"github.com/gin-gonic/gin"
)

// SetClientRouter configures all client routes
func SetClientRouter(router *gin.Engine) {
	clientRouter := router.Group("/")

	// Define client routes
	clientRouter.GET("/", func(c *gin.Context) {
		c.String(200, "This is the client route")
	})

	clientRouter.GET("/products", controllers.GetProducts)
	clientRouter.GET("/customers", controllers.GetCustomers)
	clientRouter.GET("/transactions", controllers.GetTransactions)
	clientRouter.GET("/geography", controllers.GetGeography)
}
