package controllers

import (
	"fmt"
	"net/http"
	"os/user"
	"strconv"

	"github.com/Bbanks14/dashboard-server/internal/data/database"
	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/Bbanks14/dashboard-server/internal/util/log"
	"github.com/Bbanks14/dashboard-server/pkg/helpers/params"
	"github.com/gin-gonic/gin"
)

// ClientController handles all client-facing endpoints
type ClientController struct {
	DB *database.Database
}

// NewClientController creates a new client controller instance
func NewClientController(db *database.Database) *ClientController {
	return &ClientController{DB: db}
}

// GetProducts returns all products with pagination
func (c *ClientController) GetProducts(ctx *gin.Context) {
	params := GetQueryParams(ctx)

	products, totalCount, err := c.DB.GetProducts(params)

	// Use the log package to log the error
	if err != nil {
		log.ErrorLogger.Printf("Error fetching products: %v", error)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    totalCount,
	})
}

// GetCustomers returns all customers/users with pagination
func (c *ClientController) GetCustomers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	sort := ctx.DefaultQuery("sort", "createdAt")
	search := ctx.DefaultQuery("search", "")

	users, totalCount, err := c.DB.GetUsers(page, pageSize, sort, search)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"customers": users,
		"total":     totalCount,
	})
}

// GetTransactions returns all transactions with pagination
func (c *ClientController) GetTransactions(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	sort := ctx.DefaultQuery("sort", "createdAt")
	search := ctx.DefaultQuery("search", "")

	transactions, totalCount, err := c.DB.GetTransactions(page, pageSize, sort, search)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        totalCount,
	})
}

// getGeography returns user count by location for geography charts
func (c *ClientController) GetGeography(ctx *gin.Context) {
	usersByCountry, err := c.DB.GetUsersByLocation()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": usersByCountry,
	})
}
