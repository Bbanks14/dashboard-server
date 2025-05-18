package controllers

import (
	"net/http"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/Bbanks14/dashboard-server/internal/util/errors"
	"github.com/Bbanks14/dashboard-server/pkg/helpers"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// DashboardController interface defines the contract for dashboard controllers
type DashboardController interface {
	GetUser(ctx *gin.Context)
	GetDashboardStats(ctx *gin.Context)
	GetDashboardSummary(ctx *gin.Context)
}

// dashboardController implements the DashboardController interface
type dashboardController struct {
	mongoClient      *mongo.Client
	database         *mongo.Database
	dashboardService *services.DashboardService
}

// NewDashboardController creates a new dashboard controller instance
func NewDashboardController(client *mongo.Client, adminDashboard string) DashboardController {
	db := client.Database(adminDashboard)
	return &dashboardController{
		mongoClient:      client,
		database:         db,
		dashboardService: services.NewDashboardService(client, adminDashboard),
	}
}

// GetUser retrieves data concerning current users based on date range
func (c *dashboardController) GetUser(ctx *gin.Context) { // Parse date parameters with defaults
	startDateStr := ctx.DefaultQuery("startDate", "")
	endDateStr := ctx.DefaultQuery("endDate", "")

	var startDate, endDate time.Time
	var err error

	// If no dates provided, default to last 30 days
	if startDateStr == "" || endDateStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		// Parse Hardcoded date values
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	}

	// Get Users
	users, err := c.dashboardService.GetUser(startDate, endDate)
	if err != nil {
		errors.HandleError(ctx, err, "Failed to retrieve users")
		return
	}

	// Return the user data
	ctx.JSON(http.StatusOK, gin.H{
		"users": users,
		"timeframe": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.Format("2006-01-02"),
		},
	})
}

// GetDashboardStats retrieves analytics data for the dashboard based on date range
func (c *dashboardController) GetDashboardStats(ctx *gin.Context) {
	// Parse date parameters with defaults
	startDateStr := ctx.DefaultQuery("startDate", "")
	endDateStr := ctx.DefaultQuery("endDate", "")

	var startDate, endDate time.Time
	var err error

	// If no dates provided, default to last 30 days
	if startDateStr == "" || endDateStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		// Parse Hardcoded date values
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	}

	// Get dashboard statistics
	transactions, err := c.dashboardService.GetUser(startDate, endDate)
	if err != nil {
		errors.HandleError(ctx, err, "Failed to retrieve users")
		return
	}

	stats, err := c.dashboardService.GetDashboardStats(startDate, endDate)
	if err != nil {
		errors.HandleError(ctx, err, "Failed to retrieve dashboard stats")
		return
	}

	// Return the stats data
	ctx.JSON(http.StatusOK, gin.H{
		"stats": stats,
		"timeframe": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.Format("2006-01-02"),
		},
	})
}

// GetDashboardSummary retrieves a summary of current dashboard statistics
func (c *dashboardController) GetDashboardSummary(ctx *gin.Context) {
	// Get dashboard summary
	summary, err := c.dashboardService.GetDashboardSummary()
	if err != nil {
		errors.HandleError(ctx, err, "Failed to retrieve dashboard summary")
		return
	}

	// Return the summary data
	ctx.JSON(http.StatusOK, gin.H{
		"summary":      summary,
		"generated_at": time.Now().Format(time.RFC3339),
	})
}
