package routes

import (
	"github.com/Bbanks14/dashboard-server/internal/api"
	"github.com/Bbanks14/dashboard-server/internal/auth"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	authAPI *api.AuthAPI,
	userAPI *api.UserAPI,
	dashboardAPI *api.DashboardAPI,
	jwtSecret string,
) *gin.Engine {
	router := gin.Default()

	// Public routes
	public := router.Group("/api")
	{
		public.POST("/auth/google", authAPI.GoogleLoginHandler)
		public.GET("/auth/google/callback", authAPI.GoogleCallbackHandler)
		public.GET("/auth/me", authAPI.GetCurrentUser)

		// User routes
		userAPI.RegisterRoutes(public)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(auth.GinAuthMiddleware(jwtSecret))
	{
		protected.GET("/dashboard", dashboardAPI.GetDashboard)
		// Add other protected routes...
	}

	return router
}
