package domain

import (
	"errors"
	"time"
)

type Product struct {
	ID          string    `json:"id" db:"name" binding:"required"`
	Name        string    `json:"name" db:"name" binding:"required"`
	Price       float64   `json:"price" db:"price" binding:"required,min=0"`
	Description string    `json:"description,omitempty" db:"description"`
	Category    string    `json:"category" db:"category" binding:"required"`
	Rating      float64   `json:"rating,omitempty" db:"rating" binding:"min=0,max=5"`
	Supply      int       `json:"supply,omitempty" db:"supply" binding:"min=0"`
	SKU         string    `json:"sku,omitempty" db:"sku"`
	Brand       string    `json:"brand,omitempty" db:"brand"`
	Tags        []string  `json:"tags,omitempty" db:"tags"`
	ImageURL    string    `json:"imageUrl,omitempty" db:"image_url"`
	Status      string    `json:"status" db:"status"`
	Weight      float64   `json:"weight,omitempty" db:"weight"`
	Dimensions  string    `json:"dimensions,omitempty" db:"dimensions"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type ProductStat struct {
	ID           string    `json:"id" db:"id"`
	ProductID    string    `json:"productId" db:"product_id" binding:"required"`
	Year         int       `json:"year" db:"year" binding:"required,min=2000,max=2100"`
	Month        string    `json:"month" db:"month" binding:"required"`
	TotalSales   float64   `json:"totalSales" db:"total_sales" binding:"min=0"`
	TotalUnits   int       `json:"totalUnits" db:"total_units" binding:"min=0"`
	AveragePrice float64   `json:"averagePrice,omitempty" db:"average_price"`
	Returns      int       `json:"returns,omitempty" db:"returns" binding:"min=0"`
	Revenue      float64   `json:"revenue,omitempty" db:"revenue"` // TotalSales - Returns cost
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// ProductCategory represents product categories
type ProductCategory struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" binding:"required"`
	Description string    `json:"description,omitempty" db:"description"`
	ParentID    string    `json:"parentId,omitempty" db:"parent_id"` // For nested categories
	ImageURL    string    `json:"imageUrl,omitempty" db:"image_url"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// ProductReview represents customer reviews
type ProductReview struct {
	ID         string    `json:"id" db:"id"`
	ProductID  string    `json:"productId" db:"product_id" binding:"required"`
	UserID     string    `json:"userId" db:"user_id" binding:"required"`
	Rating     float64   `json:"rating" db:"rating" binding:"required,min=1,max=5"`
	Title      string    `json:"title,omitempty" db:"title"`
	Comment    string    `json:"comment,omitempty" db:"comment"`
	IsVerified bool      `json:"isVerified" db:"is_verified"` // Verified purchase
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// ProductInventory represents inventory tracking
type ProductInventory struct {
	ID            string    `json:"id" db:"id"`
	ProductID     string    `json:"productId" db:"product_id" binding:"required"`
	Quantity      int       `json:"quantity" db:"quantity" binding:"min=0"`
	ReorderLevel  int       `json:"reorderLevel" db:"reorder_level" binding:"min=0"`
	MaxStock      int       `json:"maxStock,omitempty" db:"max_stock"`
	Location      string    `json:"location,omitempty" db:"location"` // Warehouse location
	LastRestocked time.Time `json:"lastRestocked,omitempty" db:"last_restocked"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// ProductFilter represents filters for product queries
type ProductFilter struct {
	Category   string   `json:"category,omitempty"`
	Brand      string   `json:"brand,omitempty"`
	MinPrice   *float64 `json:"minPrice,omitempty"`
	MaxPrice   *float64 `json:"maxPrice,omitempty"`
	MinRating  *float64 `json:"minRating,omitempty"`
	Status     string   `json:"status,omitempty"`
	InStock    *bool    `json:"inStock,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	SearchTerm string   `json:"searchTerm,omitempty"`
	SortBy     string   `json:"sortBy,omitempty"`    // name, price, rating, created_at
	SortOrder  string   `json:"sortOrder,omitempty"` // asc, desc
}

// ProductStatFilter represents filters for product statistics
type ProductStatFilter struct {
	ProductID string `json:"productId,omitempty"`
	Year      *int   `json:"year,omitempty"`
	Month     string `json:"month,omitempty"`
	StartDate string `json:"startDate,omitempty"` // YYYY-MM-DD
	EndDate   string `json:"endDate,omitempty"`   // YYYY-MM-DD
}

// ProductSummary represents aggregated product data for dashboards
type ProductSummary struct {
	TotalProducts   int     `json:"totalProducts"`
	TotalValue      float64 `json:"totalValue"`
	AveragePrice    float64 `json:"averagePrice"`
	AverageRating   float64 `json:"averageRating"`
	OutOfStockCount int     `json:"outOfStockCount"`
	LowStockCount   int     `json:"lowStockCount"`
	TopCategory     string  `json:"topCategory"`
	TotalCategories int     `json:"totalCategories"`
}

// ProductPerformance represents performance metrics for dashboard
type ProductPerformance struct {
	ProductID     string  `json:"productId"`
	ProductName   string  `json:"productName"`
	TotalRevenue  float64 `json:"totalRevenue"`
	TotalUnits    int     `json:"totalUnits"`
	AverageRating float64 `json:"averageRating"`
	ReviewCount   int     `json:"reviewCount"`
	Period        string  `json:"period"` // "monthly", "quarterly", "yearly"
}

// Constants for product status
const (
	ProductStatusActive       = "active"
	ProductStatusInactive     = "inactive"
	ProductStatusDiscontinued = "discontinued"
	ProductStatusOutOfStock   = "out_of_stock"
)

// Constants for sorting
const (
	SortByName      = "name"
	SortByPrice     = "price"
	SortByRating    = "rating"
	SortByCreatedAt = "created_at"
	SortByUpdatedAt = "updated_at"

	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Constants for months
var ValidMonths = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// Validation errors
var (
	ErrInvalidPrice     = errors.New("price must be greater than 0")
	ErrInvalidRating    = errors.New("rating must be between 0 and 5")
	ErrInvalidSupply    = errors.New("supply must be non-negative")
	ErrInvalidCategory  = errors.New("category is required")
	ErrInvalidProductID = errors.New("product ID is required")
	ErrInvalidMonth     = errors.New("invalid month")
	ErrInvalidYear      = errors.New("year must be between 2000 and 2100")
)

// Validate validates the product data
func (p *Product) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}

	if p.Price < 0 {
		return ErrInvalidPrice
	}

	if p.Rating < 0 || p.Rating > 5 {
		return ErrInvalidRating
	}

	if p.Supply < 0 {
		return ErrInvalidSupply
	}

	if p.Category == "" {
		return ErrInvalidCategory
	}

	return nil
}

// IsLowStock checks if product is low in stock
func (p *Product) IsLowStock(threshold int) bool {
	return p.Supply > 0 && p.Supply <= threshold
}

// IsOutOfStock checks if product is out of stock
func (p *Product) IsOutOfStock() bool {
	return p.Supply == 0
}

// CalculateRevenue calculates total revenue for the product
func (p *Product) CalculateRevenue(unitsSold int) float64 {
	return p.Price * float64(unitsSold)
}

// Validate validates the product statistics data
func (ps *ProductStat) Validate() error {
	if ps.ProductID == "" {
		return ErrInvalidProductID
	}

	if ps.Year < 2000 || ps.Year > 2100 {
		return ErrInvalidYear
	}

	if !isValidMonth(ps.Month) {
		return ErrInvalidMonth
	}

	if ps.TotalSales < 0 {
		return errors.New("total sales must be non-negative")
	}

	if ps.TotalUnits < 0 {
		return errors.New("total units must be non-negative")
	}

	return nil
}

// CalculateAveragePrice calculates average price from total sales and units
func (ps *ProductStat) CalculateAveragePrice() float64 {
	if ps.TotalUnits == 0 {
		return 0
	}
	return ps.TotalSales / float64(ps.TotalUnits)
}

// isValidMonth checks if the month string is valid
func isValidMonth(month string) bool {
	for _, validMonth := range ValidMonths {
		if month == validMonth {
			return true
		}
	}
	return false
}

// GetMonthNumber returns the month number (1-12) for a month name
func GetMonthNumber(month string) int {
	for i, validMonth := range ValidMonths {
		if month == validMonth {
			return i + 1
		}
	}
	return 0
}

// DefaultFilter returns a default product filter
func DefaultFilter() *ProductFilter {
	return &ProductFilter{
		SortBy:    SortByCreatedAt,
		SortOrder: SortOrderDesc,
	}
}

// ApplyDefaults sets default values for product
func (p *Product) ApplyDefaults() {
	if p.Status == "" {
		p.Status = ProductStatusActive
	}

	if p.Rating == 0 {
		p.Rating = 0.0 // Explicitly set to 0 for new products
	}

	p.IsActive = (p.Status == ProductStatusActive)

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
}
