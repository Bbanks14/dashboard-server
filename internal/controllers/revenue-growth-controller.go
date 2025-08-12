package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/time/rate"
)

// RevenueGrowthController handles API requests for revenue growth analysis
type RevenueGrowthController struct {
	revenueService services.RevenueGrowthService
	limiter        *rate.Limiter
}

// NewRevenueGrowthController creates a new instance of the RevenueGrowthController
func NewRevenueGrowthController(revenueService services.RevenueGrowthService) *RevenueGrowthController {
	// Allow 10 requests per second with a burst size of 20
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)
	limiter.SetBurst(20)
	return &RevenueGrowthController{
		revenueService: revenueService,
		limiter:        limiter,
	}
}

// RevenueGrowthDataRequest defines parameters for GetRevenueGrowthData
type RevenueGrowthDataRequest struct {
	baseDateRangeRequest
}

// RevenueByDateRequest defines parameters for GetRevenueByDate
type RevenueByDateRequest struct {
	baseDateRangeRequest
	Granularity string `form:"granularity" binding:"omitempty,oneof=daily weekly monthly quarterly yearly" example:"daily"`
}

// PeriodComparisonRequest defines parameters for GetPeriodSalesComparison
type PeriodComparisonRequest struct {
	CurrentStart  string `form:"currentStart" binding:"required" example:"2024-01-01"`
	CurrentEnd    string `form:"currentEnd" binding:"required" example:"2024-01-31"`
	PreviousStart string `form:"previousStart" binding:"required" example:"2023-12-01"`
	PreviousEnd   string `form:"previousEnd" binding:"required" example:"2023-12-31"`
}

// ProductRevenueSplitRequest defines parameters for GetProductRevenueSplit
type ProductRevenueSplitRequest struct {
	baseDateRangeRequest
	Limit int `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
}

// CustomerRevenueSplitRequest defines parameters for GetCustomerRevenueSplit
type CustomerRevenueSplitRequest struct {
	baseDateRangeRequest
	Limit int `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
}

// CustomerSegmentRevenueRequest defines parameters for GetCustomerSegmentRevenue
type CustomerSegmentRevenueRequest struct {
	baseDateRangeRequest
}

// RegionRevenueSplitRequest defines parameters for GetRegionRevenueSplit
type RegionRevenueSplitRequest struct {
	baseDateRangeRequest
}

// --- Route Handlers ---
// GetRevenueGrowthData retrieves complete revenue growth analysis data
func (rc *RevenueGrowthController) GetRevenueGrowthData(c *gin.Context) {
	var req RevenueGrowthDataRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetRevenueGrowthData(ctx, req.StartDate, req.EndDate)
	if err != nil {
		rc.handleServiceError(c, err, "GetRevenueGrowthData")
		return
	}
	rc.successResponse(c, data)
}

// GetRevenueByDate retrieves revenue data by date/period
func (rc *RevenueGrowthController) GetRevenueByDate(c *gin.Context) {
	var req RevenueByDateRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetRevenueByDate(ctx, req.StartDate, req.EndDate, req.Granularity)
	if err != nil {
		rc.handleServiceError(c, err, "GetRevenueByDate")
		return
	}
	rc.successResponse(c, data)
}

// GetPeriodSalesComparison retrieves sales comparison between periods
func (rc *RevenueGrowthController) GetPeriodSalesComparison(c *gin.Context) {
	var req PeriodComparisonRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	// Validate all date ranges
	if err := rc.validateComparisonDates(req); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetPeriodSalesComparison(
		ctx,
		req.CurrentStart,
		req.CurrentEnd,
		req.PreviousStart,
		req.PreviousEnd,
	)
	if err != nil {
		rc.handleServiceError(c, err, "GetPeriodSalesComparison")
		return
	}
	rc.successResponse(c, data)
}

// GetProductRevenueSplit retrieves revenue breakdown by products
func (rc *RevenueGrowthController) GetProductRevenueSplit(c *gin.Context) {
	var req ProductRevenueSplitRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetProductRevenueSplit(ctx, req.StartDate, req.EndDate, req.Limit)
	if err != nil {
		rc.handleServiceError(c, err, "GetProductRevenueSplit")
		return
	}
	rc.successResponse(c, data)
}

// GetCustomerRevenueSplit retrieves revenue breakdown by customers
func (rc *RevenueGrowthController) GetCustomerRevenueSplit(c *gin.Context) {
	var req CustomerRevenueSplitRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetCustomerRevenueSplit(ctx, req.StartDate, req.EndDate, req.Limit)
	if err != nil {
		rc.handleServiceError(c, err, "GetCustomerRevenueSplit")
		return
	}
	rc.successResponse(c, data)
}

// GetCustomerSegmentRevenue retrieves revenue breakdown by customer segments
func (rc *RevenueGrowthController) GetCustomerSegmentRevenue(c *gin.Context) {
	var req CustomerSegmentRevenueRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetCustomerSegmentRevenue(ctx, req.StartDate, req.EndDate)
	if err != nil {
		rc.handleServiceError(c, err, "GetCustomerSegmentRevenue")
		return
	}
	rc.successResponse(c, data)
}

// GetRegionRevenueSplit retrieves revenue breakdown by regions
func (rc *RevenueGrowthController) GetRegionRevenueSplit(c *gin.Context) {
	var req RegionRevenueSplitRequest
	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validateDateRange(); err != nil {
		rc.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	data, err := rc.revenueService.GetRegionRevenueSplit(ctx, req.StartDate, req.EndDate)
	if err != nil {
		rc.handleServiceError(c, err, "GetRegionRevenueSplit")
		return
	}
	rc.successResponse(c, data)
}

// --- Middleware ---
// loggingMiddleware logs details about each incoming request
func (rc *RevenueGrowthController) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Printf("[RevenueGrowth] | %d | %13v | %s | %s",
			c.Writer.Status(),
			latency,
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}

// rateLimitMiddleware applies rate limiting to incoming requests
func (rc *RevenueGrowthController) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rc.limiter.Allow() {
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
func (rc *RevenueGrowthController) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// --- Route Registration ---
// RegisterRoutes sets up all routes for the revenue growth API
func (rc *RevenueGrowthController) RegisterRoutes(r *gin.Engine) {
	// Add CORS middleware globally
	r.Use(rc.corsMiddleware())

	revenueGroup := r.Group("/revenue-growth")
	revenueGroup.Use(rc.loggingMiddleware())

	// Main data endpoints with rate limiting
	dataGroup := revenueGroup.Group("/data")
	dataGroup.Use(rc.rateLimitMiddleware())
	{
		dataGroup.GET("/complete", rc.GetRevenueGrowthData)
		dataGroup.GET("/by-date", rc.GetRevenueByDate)
		dataGroup.GET("/period-comparison", rc.GetPeriodSalesComparison)
		dataGroup.GET("/product-split", rc.GetProductRevenueSplit)
		dataGroup.GET("/customer-split", rc.GetCustomerRevenueSplit)
		dataGroup.GET("/customer-segments", rc.GetCustomerSegmentRevenue)
		dataGroup.GET("/region-split", rc.GetRegionRevenueSplit)
	}

	// Administrative endpoints
	adminGroup := revenueGroup.Group("/admin")
	{
		adminGroup.GET("/health", rc.HealthCheck)
	}

	// Add a simple test endpoint
	revenueGroup.GET("/test", func(c *gin.Context) {
		rc.successResponse(c, map[string]string{
			"message": "Revenue Growth API is working",
			"status":  "ok",
		})
	})
}

// Alternative method for RouterGroup registration
func (rc *RevenueGrowthController) RegisterRoutesWithGroup(rg *gin.RouterGroup) {
	revenueGroup := rg.Group("/revenue-growth")
	revenueGroup.Use(rc.loggingMiddleware())
	revenueGroup.Use(rc.corsMiddleware())

	// Main data endpoints with rate limiting
	dataGroup := revenueGroup.Group("/data")
	dataGroup.Use(rc.rateLimitMiddleware())
	{
		dataGroup.GET("/complete", rc.GetRevenueGrowthData)
		dataGroup.GET("/by-date", rc.GetRevenueByDate)
		dataGroup.GET("/period-comparison", rc.GetPeriodSalesComparison)
		dataGroup.GET("/product-split", rc.GetProductRevenueSplit)
		dataGroup.GET("/customer-split", rc.GetCustomerRevenueSplit)
		dataGroup.GET("/customer-segments", rc.GetCustomerSegmentRevenue)
		dataGroup.GET("/region-split", rc.GetRegionRevenueSplit)
	}

	// Administrative endpoints
	adminGroup := revenueGroup.Group("/admin")
	{
		adminGroup.GET("/health", rc.HealthCheck)
	}

	// Test endpoint
	revenueGroup.GET("/test", func(c *gin.Context) {
		rc.successResponse(c, map[string]string{
			"message": "Revenue Growth API is working",
			"status":  "ok",
		})
	})
}

// --- Helper Functions ---
func (rc *RevenueGrowthController) successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Data: data, Success: true})
}

func (rc *RevenueGrowthController) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{Error: message, Success: false})
}

func (rc *RevenueGrowthController) handleServiceError(c *gin.Context, err error, operation string) {
	log.Printf("Service error in %s: %v", operation, err)
	statusCode := http.StatusInternalServerError
	message := "An internal server error occurred"

	if err.Error() == "no data found" {
		statusCode = http.StatusNotFound
		message = "No data found for the specified criteria"
	}

	rc.errorResponse(c, statusCode, message)
}

// --- Additional Methods ---
// HealthCheck provides a health check endpoint
func (rc *RevenueGrowthController) HealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "revenue-growth-api",
		"version":   "1.0.0",
	}
	rc.successResponse(c, health)
}

// validateComparisonDates validates all dates in a period comparison request
func (rc *RevenueGrowthController) validateComparisonDates(req PeriodComparisonRequest) error {
	// Validate current period
	currentBase := baseDateRangeRequest{
		StartDate: req.CurrentStart,
		EndDate:   req.CurrentEnd,
	}
	if err := currentBase.validateDateRange(); err != nil {
		return fmt.Errorf("current period: %w", err)
	}

	// Validate previous period
	previousBase := baseDateRangeRequest{
		StartDate: req.PreviousStart,
		EndDate:   req.PreviousEnd,
	}
	if err := previousBase.validateDateRange(); err != nil {
		return fmt.Errorf("previous period: %w", err)
	}

	return nil
}
