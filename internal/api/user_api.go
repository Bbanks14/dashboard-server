package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Bbanks14/dashboard-server/internal/domain"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/gin-gonic/gin"
)

type UserAPI struct {
	userService *services.UserService
}

func NewUserAPI(userService *services.UserService) *UserAPI {
	return &UserAPI{userService: userService}
}

// RegisterRoutes sets up all user-related routes
func (a *UserAPI) RegisterRoutes(r *gin.RouterGroup) {
	userGroup := r.Group("/users")
	{
		userGroup.POST("/register", a.Register)
		userGroup.POST("/login", a.Login)
		userGroup.GET("/:id", a.GetUser)
		userGroup.PUT("/:id", a.UpdateUser)
		userGroup.PUT("/:id/password", a.UpdatePassword)
		userGroup.DELETE("/:id", a.DeleteUser)
		userGroup.GET("", a.ListUsers)
		userGroup.GET("/search", a.SearchUserByEmail)
	}
}

// In user_api.go
func (a *UserAPI) RegisterAdminRoutes(r *gin.RouterGroup) {
	adminGroup := r.Group("/admin/users")
	adminGroup.Use(adminMiddleware()) // Implement this middleware
	{
		adminGroup.GET("", a.ListUsers)
		adminGroup.PUT("/:id/role", a.UpdateUserRole)
		adminGroup.DELETE("/:id", a.DeleteUser)
	}
}

// Register handles user registration
func (a *UserAPI) Register(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := a.userService.RegisterUser(c.Request.Context(), &user); err != nil {
		handleServiceError(c, err)
		return
	}

	// Clear password before responding
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login handles user authentication
func (a *UserAPI) Login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	user, err := a.userService.AuthenticateUser(c.Request.Context(), credentials.Email, credentials.Password)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
	})
}

// GetUser retrieves a user by ID
func (a *UserAPI) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	user, err := a.userService.GetUserProfile(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles user profile updates
func (a *UserAPI) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var update domain.User
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	user, err := a.userService.UpdateUserProfile(c.Request.Context(), id, &update)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

// UpdatePassword handles password updates
func (a *UserAPI) UpdatePassword(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var passwords struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&passwords); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := a.userService.UpdatePassword(c.Request.Context(), id, passwords.CurrentPassword, passwords.NewPassword); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// DeleteUser handles user deletion
func (a *UserAPI) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	if err := a.userService.DeleteUser(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsers handles paginated user listing
func (a *UserAPI) ListUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := a.userService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize // Ceiling division

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// SearchUserByEmail handles user search by email
func (a *UserAPI) SearchUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
		return
	}

	user, err := a.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// Enhanced helper function to handle service errors
func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
			"code":  "INVALID_CREDENTIALS",
		})
	case errors.Is(err, services.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "An account with this email already exists",
			"code":  "EMAIL_EXISTS",
		})
	case errors.Is(err, services.ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please provide a valid email address",
			"code":  "INVALID_EMAIL",
		})
	case errors.Is(err, services.ErrWeakPassword):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 8 characters long and contain uppercase, lowercase, and numeric characters",
			"code":  "WEAK_PASSWORD",
		})
	case errors.Is(err, services.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input provided",
			"code":  "INVALID_INPUT",
		})
	case errors.Is(err, repositories.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
	default:
		// In production, you should log this error for debugging
		// log.Printf("Internal server error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An internal error occurred. Please try again later.",
			"code":  "INTERNAL_ERROR",
		})
	}
}

// Health check endpoint for the user service
func (a *UserAPI) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "user-api",
	})
}

// RegisterHealthRoutes adds health check routes
func (a *UserAPI) RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", a.HealthCheck)
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateUserRegistration performs additional validation for user registration
func ValidateUserRegistration(user *domain.User) []ValidationError {
	var errors []ValidationError

	if user.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name is required",
		})
	}

	if user.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is required",
		})
	}

	if user.Password == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password is required",
		})
	}

	return errors
}

// Enhanced Register with better validation
func (a *UserAPI) RegisterWithValidation(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Additional validation
	if validationErrors := ValidateUserRegistration(&user); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": validationErrors,
		})
		return
	}

	if err := a.userService.RegisterUser(c.Request.Context(), &user); err != nil {
		handleServiceError(c, err)
		return
	}

	// Clear password before responding
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// GetUserStats returns user statistics (admin endpoint)
func (a *UserAPI) GetUserStats(c *gin.Context) {
	// This would require additional service methods
	// For now, it's a placeholder for future implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "User stats endpoint - to be implemented",
	})
}

// RegisterAdminRoutes sets up admin-only routes
func (a *UserAPI) RegisterAdminRoutes(r *gin.RouterGroup) {
	adminGroup := r.Group("/admin/users")
	// Add authentication middleware here
	{
		adminGroup.GET("/stats", a.GetUserStats)
		adminGroup.GET("", a.ListUsers) // Admin can list all users
	}
}
