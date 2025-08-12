package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// DashboardController handles API requests for dashboard data.
type DashboardController struct {
	dashboardService services.DashboardService
	limiter          *rate.Limiter
}

// APIResponse defines the standard JSON response structure.
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// NewDashboardController creates a new instance of the DashboardController.
func NewDashboardController(dashboardService services.DashboardService) *DashboardController {
	// Allow 10 requests per second with a burst size of 20.
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)
	limiter.SetBurst(20)
	return &DashboardController{
		dashboardService: dashboardService,
		limiter:          limiter,
	}
}

// --- Request Binding Structs with Validation ---
// baseDateRangeRequest contains common date range fields and validation.
type baseDateRangeRequest struct {
	StartDate string `form:"startDate" binding:"required" example:"2024-01-01"`
	EndDate   string `form:"endDate" binding:"required" example:"2024-01-31"`
}

// validateDateRange performs custom validation on the date range.
func (r *baseDateRangeRequest) validateDateRange() error {
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start date format, expected YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end date format, expected YYYY-MM-DD")
	}
	if endDate.Before(startDate) {
		return fmt.Errorf("end date must be after start date")
	}
	if endDate.Sub(startDate) > (365 * 24 * time.Hour) {
		return fmt.Errorf("date range cannot exceed 1 year")
	}
	if startDate.Before(time.Now().AddDate(-5, 0, 0)) {
		return fmt.Errorf("start date cannot be more than 5 years ago")
	}
	return nil
}

// DashboardDataRequest defines the parameters for the GetDashboardData endpoint.
type DashboardDataRequest struct {
	baseDateRangeRequest
	RegionID *uuid.UUID `form:"regionID"` // Optional region filter
}

// DashboardMetricsRequest defines the parameters for the GetDashboardMetrics endpoint.
type DashboardMetricsRequest struct {
	baseDateRangeRequest
	RegionID *uuid.UUID `form:"regionID"` // Optional region filter
}

// RevenueHistoryRequest defines the parameters for the GetRevenueHistory endpoint.
type RevenueHistoryRequest struct {
	baseDateRangeRequest
}

// ConversionRatesRequest defines the parameters for the GetConversionRates endpoint.
type ConversionRatesRequest struct {
	baseDateRangeRequest
}

// ForecastDataRequest defines the parameters for the GetForecastData endpoint.
type ForecastDataRequest struct {
	baseDateRangeRequest
}

// ProductSummaryRequest defines the parameters for the GetProductSummary endpoint.
type ProductSummaryRequest struct {
	baseDateRangeRequest
}

// RegionSummaryRequest defines the parameters for the GetRegionSummary endpoint.
type RegionSummaryRequest struct {
	baseDateRangeRequest
}

// --- Route Handlers ---
// GetDashboardData retrieves complete dashboard data including all metrics and summaries.
func (dc *DashboardController) GetDashboardData(c *gin.Context) {
	var req DashboardDataRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := dc.dashboardService.GetDashboardData(ctx, req.StartDate, req.EndDate, req.RegionID)
	if err != nil {
		dc.handleServiceError(c, err, "GetDashboardData")
		return
	}
	dc.successResponse(c, data)
}

// GetDashboardMetrics retrieves key dashboard performance metrics.
func (dc *DashboardController) GetDashboardMetrics(c *gin.Context) {
	var req DashboardMetricsRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	metrics, err := dc.dashboardService.GetDashboardMetrics(ctx, req.StartDate, req.EndDate, req.RegionID)
	if err != nil {
		dc.handleServiceError(c, err, "GetDashboardMetrics")
		return
	}
	dc.successResponse(c, metrics)
}

// GetRevenueHistory retrieves historical revenue data for charting.
func (dc *DashboardController) GetRevenueHistory(c *gin.Context) {
	var req RevenueHistoryRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	history, err := dc.dashboardService.GetRevenueHistory(ctx, req.StartDate, req.EndDate)
	if err != nil {
		dc.handleServiceError(c, err, "GetRevenueHistory")
		return
	}
	dc.successResponse(c, history)
}

// GetConversionRates retrieves sales conversion rate data.
func (dc *DashboardController) GetConversionRates(c *gin.Context) {
	var req ConversionRatesRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	rates, err := dc.dashboardService.GetConversionRates(ctx, req.StartDate, req.EndDate)
	if err != nil {
		dc.handleServiceError(c, err, "GetConversionRates")
		return
	}
	dc.successResponse(c, rates)
}

// GetForecastData retrieves sales forecast data.
func (dc *DashboardController) GetForecastData(c *gin.Context) {
	var req ForecastDataRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	forecasts, err := dc.dashboardService.GetForecastData(ctx, req.StartDate, req.EndDate)
	if err != nil {
		dc.handleServiceError(c, err, "GetForecastData")
		return
	}
	dc.successResponse(c, forecasts)
}

// GetProductSummary retrieves product performance summary.
func (dc *DashboardController) GetProductSummary(c *gin.Context) {
	var req ProductSummaryRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	summary, err := dc.dashboardService.GetProductSummary(ctx, req.StartDate, req.EndDate)
	if err != nil {
		dc.handleServiceError(c, err, "GetProductSummary")
		return
	}
	dc.successResponse(c, summary)
}

// GetRegionSummary retrieves regional performance summary.
func (dc *DashboardController) GetRegionSummary(c *gin.Context) {
	var req RegionSummaryRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		dc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	summary, err := dc.dashboardService.GetRegionSummary(ctx, req.StartDate, req.EndDate)
	if err != nil {
		dc.handleServiceError(c, err, "GetRegionSummary")
		return
	}
	dc.successResponse(c, summary)
}

// --- Middleware ---
// loggingMiddleware logs details about each incoming request.
func (dc *DashboardController) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Printf("[Dashboard] | %d | %13v | %s | %s",
			c.Writer.Status(),
			latency,
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}

// rateLimitMiddleware applies rate limiting to incoming requests.
func (dc *DashboardController) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !dc.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, APIResponse{
				Error:   "Too many requests",
				Success: false,
			})
			return
		}
		c.Next()
	}
}

// corsMiddleware handles CORS headers for cross-origin requests
func (dc *DashboardController) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// --- Route Registration ---
// RegisterRoutes sets up all the routes for the dashboard API.
func (dc *DashboardController) RegisterRoutes(r *gin.Engine) {
	// Add CORS middleware globally
	r.Use(dc.corsMiddleware())

	dashboardGroup := r.Group("/dashboard")
	dashboardGroup.Use(dc.loggingMiddleware())

	// Main data endpoints with rate limiting
	dataGroup := dashboardGroup.Group("/data")
	dataGroup.Use(dc.rateLimitMiddleware())
	{
		// Complete dashboard data - single endpoint for full dashboard
		dataGroup.GET("/complete", dc.GetDashboardData)
		// Individual data endpoints for specific dashboard components
		dataGroup.GET("/metrics", dc.GetDashboardMetrics)
		dataGroup.GET("/revenue-history", dc.GetRevenueHistory)
		dataGroup.GET("/conversion-rates", dc.GetConversionRates)
		dataGroup.GET("/forecasts", dc.GetForecastData)
		dataGroup.GET("/products", dc.GetProductSummary)
		dataGroup.GET("/regions", dc.GetRegionSummary)
	}

	// Administrative endpoints (no rate limiting for health checks)
	adminGroup := dashboardGroup.Group("/admin")
	{
		adminGroup.GET("/cache-stats", dc.GetCacheStats)
		adminGroup.POST("/invalidate-cache", dc.InvalidateCache)
		adminGroup.GET("/health", dc.HealthCheck)
	}

	// Add a simple test endpoint that doesn't require any parameters
	dashboardGroup.GET("/test", func(c *gin.Context) {
		dc.successResponse(c, map[string]string{
			"message": "Dashboard API is working",
			"status":  "ok",
		})
	})
}

// Alternative method for RouterGroup registration (for flexibility)
func (dc *DashboardController) RegisterRoutesWithGroup(rg *gin.RouterGroup) {
	dashboardGroup := rg.Group("/dashboard")
	// Apply logging and CORS middleware
	dashboardGroup.Use(dc.loggingMiddleware())
	dashboardGroup.Use(dc.corsMiddleware())

	// Main data endpoints with rate limiting
	dataGroup := dashboardGroup.Group("/data")
	dataGroup.Use(dc.rateLimitMiddleware())
	{
		dataGroup.GET("/complete", dc.GetDashboardData)
		dataGroup.GET("/metrics", dc.GetDashboardMetrics)
		dataGroup.GET("/revenue-history", dc.GetRevenueHistory)
		dataGroup.GET("/conversion-rates", dc.GetConversionRates)
		dataGroup.GET("/forecasts", dc.GetForecastData)
		dataGroup.GET("/products", dc.GetProductSummary)
		dataGroup.GET("/regions", dc.GetRegionSummary)
	}

	// Administrative endpoints
	adminGroup := dashboardGroup.Group("/admin")
	{
		adminGroup.GET("/cache-stats", dc.GetCacheStats)
		adminGroup.POST("/invalidate-cache", dc.InvalidateCache)
		adminGroup.GET("/health", dc.HealthCheck)
	}

	// Test endpoint
	dashboardGroup.GET("/test", func(c *gin.Context) {
		dc.successResponse(c, map[string]string{
			"message": "Dashboard API is working",
			"status":  "ok",
		})
	})
}

// --- Helper Functions ---
func (dc *DashboardController) successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Data: data, Success: true})
}

func (dc *DashboardController) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{Error: message, Success: false})
}

func (dc *DashboardController) handleServiceError(c *gin.Context, err error, operation string) {
	log.Printf("Service error in %s: %v", operation, err)
	// You could enhance this with specific error type handling
	// For example, check if it's a validation error, not found error, etc.
	statusCode := http.StatusInternalServerError
	message := "An internal server error occurred"

	// Example of specific error handling:
	if err.Error() == "no data found" {
		statusCode = http.StatusNotFound
		message = "No data found for the specified criteria"
	}

	dc.errorResponse(c, statusCode, message)
}

// --- Additional Handler Methods ---
// GetCacheStats returns cache statistics (placeholder implementation).
func (dc *DashboardController) GetCacheStats(c *gin.Context) {
	// In a real implementation, you would get actual cache stats
	stats := map[string]interface{}{
		"status":     "active",
		"hit_rate":   0.85,
		"total_keys": 1250,
		"memory_mb":  45.2,
	}
	dc.successResponse(c, stats)
}

// InvalidateCache clears the dashboard data cache.
func (dc *DashboardController) InvalidateCache(c *gin.Context) {
	// In a real implementation, you would trigger cache invalidation
	// This might be implemented in your service layer
	dc.successResponse(c, map[string]string{
		"status":  "success",
		"message": "Dashboard cache invalidated",
	})
}

// HealthCheck provides a health check endpoint for the dashboard service.
func (dc *DashboardController) HealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "dashboard-api",
		"version":   "1.0.0",
	}
	dc.successResponse(c, health)
}
