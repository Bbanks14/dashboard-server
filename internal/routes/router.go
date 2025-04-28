// File: internal/routes/router.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/dashboard-backend/internal/controllers"
	"github.com/yourusername/dashboard-backend/internal/data"
	"github.com/yourusername/dashboard-backend/internal/middleware"
	"github.com/yourusername/dashboard-backend/internal/services"
	"github.com/yourusername/dashboard-backend/internal/util/config"
	"github.com/yourusername/dashboard-backend/internal/util/logger"
)

// SetupRouter configures and returns the application router
func SetupRouter(db *data.Database, logger *logger.Logger, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.NewLoggerMiddleware(logger).LogRequest())
	router.Use(middleware.NewCorsMiddleware(cfg).EnableCORS())

	// Initialize repositories
	userRepo := data.NewUserRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo)
	dashboardService := services.NewDashboardService(db)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	dashboardController := controllers.NewDashboardController(dashboardService)

	// Initialize middleware that requires services
	authMiddleware := middleware.NewAuthMiddleware(authService, cfg)

	// Setup routes
	SetupAuthRoutes(router, authController)
	SetupUserRoutes(router, userController, authMiddleware)
	SetupDashboardRoutes(router, dashboardController, authMiddleware)

	return router
}
