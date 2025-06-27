package controllers

import (
	"net/http"
	"strconv"

	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/Bbanks14/dashboard-server/internal/structs"
	"github.com/gin-gonic/gin"
)

// ClientController handles all client-facing endpoints
type ClientController struct {
	clientService services.ClientServiceInterface
}

// NewClientController creates a new client controller instance
func NewClientController(clientService services.ClientServiceInterface) *ClientController {
	return &ClientController{
		clientService: clientService,
	}
}

// GetClients handles GET /api/clients - returns client list
func (c *ClientController) GetClients(ctx *gin.Context) {
	params, ok := getQueryParams(ctx)
	if !ok {
		return
	}

	clients, totalCount, err := c.clientService.GetClients(params.Page, params.PageSize, params.Sort, params.Search)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"clients": clients,
		"total":   totalCount,
	})
}

// GetProducts returns all products with pagination
func (c *ClientController) GetProducts(ctx *gin.Context) {
	params, ok := getQueryParams(ctx)
	if !ok {
		return
	}

	products, totalCount, err := c.clientService.GetProducts(params.Page, params.PageSize, params.Sort, params.Search)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    totalCount,
	})
}

// GetCustomers returns all customers/users with pagination
func (c *ClientController) GetCustomers(ctx *gin.Context) {
	params, ok := getQueryParams(ctx)
	if !ok {
		return
	}

	customers, totalCount, err := c.clientService.GetCustomers(params.Page, params.PageSize, params.Sort, params.Search)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"total":     totalCount,
	})
}

// GetTransactions returns all transactions with pagination
func (c *ClientController) GetTransactions(ctx *gin.Context) {
	params, ok := getQueryParams(ctx)
	if !ok {
		return
	}

	transactions, totalCount, err := c.clientService.GetTransactions(params.Page, params.PageSize, params.Sort, params.Search)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        totalCount,
	})
}

// GetGeography returns user count by location for geography charts
func (c *ClientController) GetGeography(ctx *gin.Context) {
	data, err := c.clientService.GetGeography()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve geography data"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// getQueryParams extracts common query parameters from request
func getQueryParams(ctx *gin.Context) (structs.QueryParams, bool) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return structs.QueryParams{}, false
	}

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pageSize parameter"})
		return structs.QueryParams{}, false
	}

	sort := ctx.DefaultQuery("sort", "id")
	search := ctx.Query("search")

	return structs.QueryParams{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Search:   search,
	}, true
}
