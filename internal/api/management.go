package api

import (
	"github.com/Bbanks14/dashboard-server/internal/controllers"

	"github.com/gin-gonic/gin"
)

// SetManagementRouter configures all management routes
func SetManagementRouter(router *gin.Engine) {
	managementRouter := router.Group("/")

	// Define management routes
	managementRouter.GET("/", func(c *gin.Context) {
		c.String(200, "This is the management route")
	})

	managementRouter.GET("/admins", controllers.GetAdmins)
	managementRouter.GET("/performace/:id", controllers.GetUserPerformance)
}
