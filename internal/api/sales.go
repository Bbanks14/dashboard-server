package api

import (
	"github.com/Bbanks14/dashboard-server/internal/controllers"

	"github.com/gin-gonic/gin"
)

// SetSalesRouter configures all sales routes
func SetSalesRouter(router *gin.Engine) {
	salesRouter := router.Group("/")

	// Define sales routes
	salesRouter.GET("/", func(c *gin.Context) {
		c.String(200, "This is the sales endpoint")
	})

	salesRouter.GET("/sales", controllers.GetSales)
}
