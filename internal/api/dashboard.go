package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/controllers"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"
)

// GoogleUser represents user info from Google OAuth
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// DashboardAPI encapsulates all dashboard-related API components
type DashboardAPI struct {
	DB          *pgxpool.Pool
	Router      *gin.Engine
	Controller  *controllers.DashboardController
	RateLimiter *rate.Limiter
	OAuthConfig *oauth2.Config
	JWTSecret   []byte
	StateStore  map[string]time.Time // In production, use Redis or database
}

// NewDashboardAPI creates a new DashboardAPI instance
func NewDashboardAPI(db *pgxpool.Pool) *DashboardAPI {
	// Initialize repository
	dashboardRepo, err := repositories.NewDashboardRepository(db)
	if err != nil {
		log.Fatalf("Failed to create dashboard repository: %v", err)
	}

	// Initialize service
	dashboardService := services.NewDashboardService(dashboardRepo)

	// Initialize controller
	dashboardController := controllers.NewDashboardController(dashboardService)

	// Create rate limiter (10 req/sec with burst of 20)
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)
	limiter.SetBurst(20)

	// Setup Google OAuth
	oauthConfig := &oauth2.Config{
		ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// JWT secret
	jwtSecret := []byte(getEnv("JWT_SECRET", "your-secret-key-change-in-production"))

	return &DashboardAPI{
		DB:          db,
		Controller:  dashboardController,
		RateLimiter: limiter,
		OAuthConfig: oauthConfig,
		JWTSecret:   jwtSecret,
		StateStore:  make(map[string]time.Time),
	}
}

// SetupRoutes configures all API routes
func (api *DashboardAPI) SetupRoutes() {
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://yourdomain.com"}, // Configure for your frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Global middleware
	router.Use(api.loggingMiddleware())
	router.Use(api.cleanupStateMiddleware()) // Clean expired states

	// Public routes (no auth required)
	public := router.Group("/")
	{
		public.GET("/health", api.healthCheck)
		public.GET("/auth/google/login", api.googleLogin)
		public.GET("/auth/google/callback", api.googleCallback)
	}

	// Apply global auth middleware to protected routes
	router.Use(func(c *gin.Context) {
		// Skip auth for public routes
		if strings.HasPrefix(c.Request.URL.Path, "/health") ||
			strings.HasPrefix(c.Request.URL.Path, "/auth/") {
			c.Next()
			return
		}
		// Apply auth middleware for all other routes
		api.authMiddleware()(c)
	})

	// Apply rate limiting to dashboard routes
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/dashboard") {
			api.rateLimitMiddleware()(c)
		}
		c.Next()
	})

	// Register dashboard routes (your controller handles the /dashboard prefix)
	api.Controller.RegisterRoutes(router)

	api.Router = router
}

// StartServer runs the HTTP server
func (api *DashboardAPI) StartServer(port string) {
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      api.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// --- Auth Handlers ---

func (api *DashboardAPI) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

func (api *DashboardAPI) googleLogin(c *gin.Context) {
	// Generate state token
	state := api.generateState()
	api.StateStore[state] = time.Now().Add(10 * time.Minute) // 10 min expiry

	// Get OAuth URL
	url := api.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": url,
	})
}

func (api *DashboardAPI) googleCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	if !api.verifyState(state) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := api.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to exchange authorization code",
		})
		return
	}

	// Get user info
	client := api.OAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user info",
		})
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		log.Printf("Failed to decode user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode user info",
		})
		return
	}

	// Generate JWT token
	jwtToken, err := api.generateJWT(googleUser)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	// Clean up state
	delete(api.StateStore, state)

	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"user": gin.H{
			"id":      googleUser.ID,
			"email":   googleUser.Email,
			"name":    googleUser.Name,
			"picture": googleUser.Picture,
		},
	})
}

// --- Middleware Implementations ---

func (api *DashboardAPI) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Printf("[%s] %s | %d | %s | %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
		)
	}
}

func (api *DashboardAPI) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
			})
			return
		}

		// Validate JWT
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return api.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(*JWTClaims); ok {
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("name", claims.Name)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		c.Next()
	}
}

func (api *DashboardAPI) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !api.RateLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}
		c.Next()
	}
}

func (api *DashboardAPI) cleanupStateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clean expired states
		now := time.Now()
		for state, expiry := range api.StateStore {
			if now.After(expiry) {
				delete(api.StateStore, state)
			}
		}
		c.Next()
	}
}

// --- Helper Functions ---

func (api *DashboardAPI) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (api *DashboardAPI) verifyState(state string) bool {
	expiry, exists := api.StateStore[state]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

func (api *DashboardAPI) generateJWT(user GoogleUser) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dashboard-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(api.JWTSecret)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
