package api

import (
	"github.com/Bbanks14/internal/repositories/controllers"
	"github.com/gin-gonic/gin"
)

func AddSearchRoutes(r *gin.Engine) {
	auth := r.Group("/")
	auth.Use(middleware, AuthMiddleware())

	auth.GET("/search", controllers.Search)
	auth.GET("/sync-sales", controllers.SynSales)
}
