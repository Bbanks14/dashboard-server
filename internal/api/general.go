package api

import (
	"github.com/Bbanks14/dashboard-server/internal/controllers"

	"github.com/gin-gonic/gin"
)

// SetGeneralRouter configures all general routes
func SetGeneralRouter(router *gin.Engine) {
	generalRouter := router.Group("/")

	generalRouter.GET("/user/:id", controllers.GetUser)
}
