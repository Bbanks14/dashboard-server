// File: internal/controllers/user_controller.go
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/dashboard-backend/internal/models"
	"github.com/yourusername/dashboard-backend/internal/services"
	"github.com/yourusername/dashboard-backend/pkg/helpers"
)

// UserController handles HTTP requests for user operations
type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new user controller
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetUser handles requests to get a user by ID
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		helpers.ErrorResponse(ctx, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := c.userService.GetUser(id)
	if err != nil {
		helpers.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		helpers.ErrorResponse(ctx, http.StatusNotFound, "User not found")
		return
	}

	helpers.SuccessResponse(ctx, http.StatusOK, user)
}

// CreateUser handles requests to create a new user
func (c *UserController) CreateUser(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		helpers.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := c.userService.CreateUser(&user); err != nil {
		helpers.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create user")
		return
	}

	helpers.SuccessResponse(ctx, http.StatusCreated, user)
}
