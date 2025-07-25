package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/Bbanks14/dashboard-server/internal/structs"
	"github.com/google/uuid"
)

// NewProductService creates a new product service instance
func NewProductService(productRepo *repositories.ProductRepository, logger *log.Logger) *structs.ProductService {
	if logger == nil {
		logger = log.Default()
	}

	return &structs.ProductService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// GetAllProducts retrieves all products with their statistics
func (ps *ProductService) GetAllProducts(ctx context.Context) ([]ProductResponse, error) {
	products, err := ps.productRepo.GetAllProducts(ctx)
	if err != nil {
		ps.logger.Printf("Error getting all products: %v", err)
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}

	return ps.convertToProductResponses(products), nil
}

// GetProductByID retrieves a single product by ID
func (ps *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*ProductResponse, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	product, err := ps.productRepo.GetProductByID(ctx, id)
	if err != nil {
		ps.logger.Printf("Error getting product by ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}

	response := ps.convertToProductResponse(*product)
	return &response, nil
}

// CreateProduct creates a new product
func (ps *ProductService) CreateProduct(ctx context.Context, req ProductRequest) (*ProductResponse, error) {
	if err := ps.validateProductRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	product := repositories.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Rating:      req.Rating,
		Category:    req.Category,
		Supply:      req.Supply,
	}

	err := ps.productRepo.CreateProduct(ctx, &product)
	if err != nil {
		ps.logger.Printf("Error creating product: %v", err)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Initialize default statistics for the product
	currentYear := time.Now().Year()
	defaultStat := repositories.ProductStat{
		ProductID:            product.ID,
		YearlySalesTotal:     0,
		YearlyTotalSoldUnits: 0,
		Year:                 currentYear,
		MonthlyData:          []repositories.MonthlyData{},
		DailyData:            []repositories.DailyData{},
	}

	if err := ps.productRepo.CreateProductStat(ctx, &defaultStat); err != nil {
		ps.logger.Printf("Warning: Failed to create default stats for product %s: %v", product.ID, err)
	}

	// Retrieve the created product with stats
	createdProduct, err := ps.productRepo.GetProductByID(ctx, product.ID)
	if err != nil {
		ps.logger.Printf("Error retrieving created product: %v", err)
		return nil, fmt.Errorf("failed to retrieve created product: %w", err)
	}

	response := ps.convertToProductResponse(*createdProduct)
	return &response, nil
}

// UpdateProduct updates an existing product
func (ps *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req ProductRequest) (*ProductResponse, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	if err := ps.validateProductRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if product exists
	_, err := ps.productRepo.GetProductByID(ctx, id)
	if err != nil {
		ps.logger.Printf("Error checking product existence: %v", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	product := repositories.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Rating:      req.Rating,
		Category:    req.Category,
		Supply:      req.Supply,
	}

	err = ps.productRepo.UpdateProduct(ctx, id, &product)
	if err != nil {
		ps.logger.Printf("Error updating product %s: %v", id, err)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Retrieve the updated product
	updatedProduct, err := ps.productRepo.GetProductByID(ctx, id)
	if err != nil {
		ps.logger.Printf("Error retrieving updated product: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated product: %w", err)
	}

	response := ps.convertToProductResponse(*updatedProduct)
	return &response, nil
}

// DeleteProduct deletes a product by ID
func (ps *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid product ID")
	}

	err := ps.productRepo.DeleteProduct(ctx, id)
	if err != nil {
		ps.logger.Printf("Error deleting product %s: %v", id, err)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// GetProductsByCategory retrieves products by category
func (ps *ProductService) GetProductsByCategory(ctx context.Context, category string) ([]ProductResponse, error) {
	if category == "" {
		return nil, errors.New("category cannot be empty")
	}

	products, err := ps.productRepo.GetProductsByCategory(ctx, category)
	if err != nil {
		ps.logger.Printf("Error getting products by category %s: %v", category, err)
		return nil, fmt.Errorf("failed to retrieve products by category: %w", err)
	}

	return ps.convertToProductResponses(products), nil
}

// SearchProducts searches for products by name or description
func (ps *ProductService) SearchProducts(ctx context.Context, query string) ([]ProductResponse, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	products, err := ps.productRepo.SearchProducts(ctx, query)
	if err != nil {
		ps.logger.Printf("Error searching products with query %s: %v", query, err)
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return ps.convertToProductResponses(products), nil
}

// GetProductsWithPagination retrieves products with pagination
func (ps *ProductService) GetProductsWithPagination(ctx context.Context, page, pageSize int) (*ProductListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	products, totalCount, err := ps.productRepo.GetProductsWithPagination(ctx, page, pageSize)
	if err != nil {
		ps.logger.Printf("Error getting products with pagination: %v", err)
		return nil, fmt.Errorf("failed to retrieve products with pagination: %w", err)
	}

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &ProductListResponse{
		Products:    ps.convertToProductResponses(products),
		TotalCount:  totalCount,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

// UpdateProductSupply updates the supply of a product
func (ps *ProductService) UpdateProductSupply(ctx context.Context, id uuid.UUID, newSupply int) error {
	if id == uuid.Nil {
		return errors.New("invalid product ID")
	}

	if newSupply < 0 {
		return errors.New("supply cannot be negative")
	}

	err := ps.productRepo.UpdateProductSupply(ctx, id, newSupply)
	if err != nil {
		ps.logger.Printf("Error updating product supply for %s: %v", id, err)
		return fmt.Errorf("failed to update product supply: %w", err)
	}

	return nil
}

// UpdateProductStats updates product statistics
func (ps *ProductService) UpdateProductStats(ctx context.Context, productID uuid.UUID, req ProductStatRequest) error {
	if productID == uuid.Nil {
		return errors.New("invalid product ID")
	}

	if req.Year < 2000 || req.Year > time.Now().Year()+1 {
		return errors.New("invalid year")
	}

	stat := repositories.ProductStat{
		ProductID:            productID,
		YearlySalesTotal:     req.YearlySalesTotal,
		YearlyTotalSoldUnits: req.YearlyTotalSoldUnits,
		Year:                 req.Year,
		MonthlyData:          req.MonthlyData,
		DailyData:            req.DailyData,
	}

	err := ps.productRepo.UpdateProductStat(ctx, productID, &stat)
	if err != nil {
		ps.logger.Printf("Error updating product stats for %s: %v", productID, err)
		return fmt.Errorf("failed to update product statistics: %w", err)
	}

	return nil
}

// ProcessSaleTransaction processes a sale transaction and updates inventory and stats
func (ps *ProductService) ProcessSaleTransaction(ctx context.Context, transaction SaleTransaction) error {
	if transaction.ProductID == uuid.Nil {
		return errors.New("invalid product ID")
	}

	if transaction.UnitsSold <= 0 {
		return errors.New("units sold must be positive")
	}

	if transaction.SaleAmount < 0 {
		return errors.New("sale amount cannot be negative")
	}

	if transaction.TransactionDate.IsZero() {
		transaction.TransactionDate = time.Now()
	}

	// Get current product
	product, err := ps.productRepo.GetProductByID(ctx, transaction.ProductID)
	if err != nil {
		ps.logger.Printf("Error getting product for sale transaction: %v", err)
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Check if sufficient inventory
	if product.Supply < transaction.UnitsSold {
		return fmt.Errorf("insufficient inventory: available=%d, requested=%d", product.Supply, transaction.UnitsSold)
	}

	// Update inventory
	newSupply := product.Supply - transaction.UnitsSold
	err = ps.productRepo.UpdateProductSupply(ctx, transaction.ProductID, newSupply)
	if err != nil {
		ps.logger.Printf("Error updating inventory for sale: %v", err)
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	// Update statistics
	updatedStat := product.Stat
	updatedStat.YearlySalesTotal += transaction.SaleAmount
	updatedStat.YearlyTotalSoldUnits += transaction.UnitsSold

	// Update monthly data
	month := int(transaction.TransactionDate.Month())
	monthlyFound := false
	for i, monthData := range updatedStat.MonthlyData {
		if monthData.Month == month {
			updatedStat.MonthlyData[i].TotalSales += transaction.SaleAmount
			updatedStat.MonthlyData[i].TotalUnits += transaction.UnitsSold
			monthlyFound = true
			break
		}
	}
	if !monthlyFound {
		updatedStat.MonthlyData = append(updatedStat.MonthlyData, repositories.MonthlyData{
			Month:      month,
			TotalSales: transaction.SaleAmount,
			TotalUnits: transaction.UnitsSold,
		})
	}

	// Update daily data (keep only last 30 days)
	transactionDate := transaction.TransactionDate.Truncate(24 * time.Hour)
	dailyFound := false
	for i, dailyData := range updatedStat.DailyData {
		if dailyData.Date.Equal(transactionDate) {
			updatedStat.DailyData[i].TotalSales += transaction.SaleAmount
			updatedStat.DailyData[i].TotalUnits += transaction.UnitsSold
			dailyFound = true
			break
		}
	}
	if !dailyFound {
		updatedStat.DailyData = append(updatedStat.DailyData, repositories.DailyData{
			Date:       transactionDate,
			TotalSales: transaction.SaleAmount,
			TotalUnits: transaction.UnitsSold,
		})
	}

	// Keep only last 30 days of daily data
	cutoffDate := time.Now().AddDate(0, 0, -30)
	var filteredDailyData []repositories.DailyData
	for _, dailyData := range updatedStat.DailyData {
		if dailyData.Date.After(cutoffDate) {
			filteredDailyData = append(filteredDailyData, dailyData)
		}
	}
	updatedStat.DailyData = filteredDailyData

	err = ps.productRepo.UpdateProductStat(ctx, transaction.ProductID, &updatedStat)
	if err != nil {
		ps.logger.Printf("Error updating product stats for sale: %v", err)
		return fmt.Errorf("failed to update product statistics: %w", err)
	}

	ps.logger.Printf("Sale transaction processed successfully for product %s: %d units sold for $%.2f",
		transaction.ProductID, transaction.UnitsSold, transaction.SaleAmount)

	return nil
}

// GetLowStockProducts retrieves products with low stock (below threshold)
func (ps *ProductService) GetLowStockProducts(ctx context.Context, threshold int) ([]ProductResponse, error) {
	if threshold < 0 {
		threshold = 10 // Default threshold
	}

	products, err := ps.productRepo.GetAllProducts(ctx)
	if err != nil {
		ps.logger.Printf("Error getting products for low stock check: %v", err)
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}

	var lowStockProducts []repositories.Product
	for _, product := range products {
		if product.Supply <= threshold {
			lowStockProducts = append(lowStockProducts, product)
		}
	}

	return ps.convertToProductResponses(lowStockProducts), nil
}

// GetTopSellingProducts retrieves top-selling products based on yearly sales
func (ps *ProductService) GetTopSellingProducts(ctx context.Context, limit int) ([]ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	products, err := ps.productRepo.GetAllProducts(ctx)
	if err != nil {
		ps.logger.Printf("Error getting products for top selling: %v", err)
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}

	// Sort by yearly sales total (descending)
	for i := 0; i < len(products)-1; i++ {
		for j := i + 1; j < len(products); j++ {
			if products[i].Stat.YearlySalesTotal < products[j].Stat.YearlySalesTotal {
				products[i], products[j] = products[j], products[i]
			}
		}
	}

	// Limit results
	if len(products) > limit {
		products = products[:limit]
	}

	return ps.convertToProductResponses(products), nil
}

// Helper methods

func (ps *ProductService) validateProductRequest(req ProductRequest) error {
	if req.Name == "" {
		return errors.New("product name is required")
	}
	if len(req.Name) > 255 {
		return errors.New("product name too long")
	}
	if req.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if req.Rating < 0 || req.Rating > 5 {
		return errors.New("rating must be between 0 and 5")
	}
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Supply < 0 {
		return errors.New("supply cannot be negative")
	}
	return nil
}

func (ps *ProductService) convertToProductResponse(product repositories.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Rating:      product.Rating,
		Category:    product.Category,
		Supply:      product.Supply,
		Stat:        product.Stat,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func (ps *ProductService) convertToProductResponses(products []repositories.Product) []ProductResponse {
	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = ps.convertToProductResponse(product)
	}
	return responses
}
