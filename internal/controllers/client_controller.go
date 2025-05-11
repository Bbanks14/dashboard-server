package controllers

import (
	"net/http"
	"strconv"

	"github.com/Bbanks14/dashboard-server/internal/data/database"
	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/Bbanks14/dashboard-server/internal/structs"
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
	params, ok := GetQueryParams(ctx)
	if !ok {
		return
	}

	products, totalCount, err := c.DB.GetProducts(params.Page, params.PageSize, params.Sort, params.Search)

	ctx.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    totalCount,
	})
}

// GetCustomers returns all customers/users with pagination
func (c *ClientController) GetCustomers(ctx *gin.Context) {
	params, ok := GetQueryParams(ctx)
	if !ok {
		return
	}
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
	params, ok := GetQueryParams(ctx)
	if !ok {
		return
	}
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
