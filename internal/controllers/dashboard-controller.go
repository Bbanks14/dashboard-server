// File: internal/controllers/dashboard_controller.go
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/dashboard-backend/internal/services"
	"github.com/yourusername/dashboard-backend/pkg/helpers"
)

// DashboardController handles HTTP requests for dashboard operations
type DashboardController struct {
	dashboardService *services.DashboardService
}

// NewDashboardController creates a new dashboard controller
func NewDashboardController(dashboardService *services.DashboardService) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

// GetDashboardStats handles requests to get dashboard stats within a date range
func (c *DashboardController) GetDashboardStats(ctx *gin.Context) {
	// Parse query parameters for date range
	startDateStr := ctx.DefaultQuery("start_date", "")
	endDateStr := ctx.DefaultQuery("end_date", "")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			helpers.ErrorResponse(ctx, http.StatusBadRequest, "Invalid start date format")
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			helpers.ErrorResponse(ctx, http.StatusBadRequest, "Invalid end date format")
			return
		}
	} else {
		// Default to current date
		endDate = time.Now()
	}

	stats, err := c.dashboardService.GetDashboardStats(startDate, endDate)
	if err != nil {
		helpers.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get dashboard stats")
		return
	}

	helpers.SuccessResponse(ctx, http.StatusOK, stats)
}

// GetDashboardSummary handles requests to get a summary of current dashboard stats
func (c *DashboardController) GetDashboardSummary(ctx *gin.Context) {
	summary, err := c.dashboardService.GetDashboardSummary()
	if err != nil {
		helpers.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get dashboard summary")
		return
	}

	helpers.SuccessResponse(ctx, http.StatusOK, summary)
}
