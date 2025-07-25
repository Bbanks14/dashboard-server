package structs

import (
	"log"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/google/uuid"
)

// ProductStruct represents a product entity
type ProductStruct struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"price" json:"name"`
	Price       int64     `db:"price" json:"price"`
	Description string    `db:"description" json:"description"`
	Category    string    `db:"category" json:"category"`
	Rating      int64     `db:"rating" json:"rating"`
	Supply      int64     `db:"supply" json:"supply"`
	StartsAt    time.Time `db:"starts_at" json:"StartsAt"`
	EndsAt      time.Time `db:"ends_at" json:"endsAt"`
}

// ProductService handles business logic for products
type ProductService struct {
	productRepo *repositories.ProductRepository
	logger      *log.Logger
}

// ProductRequest represents a request to create or update a product
type ProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description" validate:"max=1000"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Rating      float64 `json:"rating" validate:"min=0,max=5"`
	Category    string  `json:"category" validate:"required,min=1,max=100"`
	Supply      int     `json:"supply" validate:"min=0"`
}

// ProductResponse represents the response structure for products
type ProductResponse struct {
	ID          uuid.UUID                `json:"_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Price       float64                  `json:"price"`
	Rating      float64                  `json:"rating"`
	Category    string                   `json:"category"`
	Supply      int                      `json:"supply"`
	Stat        repositories.ProductStat `json:"stat"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
}

// ProductListResponse represents paginated product list response
type ProductListResponse struct {
	Products    []ProductResponse `json:"products"`
	TotalCount  int64             `json:"totalCount"`
	CurrentPage int               `json:"currentPage"`
	PageSize    int               `json:"pageSize"`
	TotalPages  int               `json:"totalPages"`
}

// ProductStatRequest represents a request to update product statistics
type ProductStatRequest struct {
	YearlySalesTotal     float64                    `json:"yearlySalesTotal" validate:"min=0"`
	YearlyTotalSoldUnits int                        `json:"yearlyTotalSoldUnits" validate:"min=0"`
	Year                 int                        `json:"year" validate:"required"`
	MonthlyData          []repositories.MonthlyData `json:"monthlyData"`
	DailyData            []repositories.DailyData   `json:"dailyData"`
}

// SaleTransaction represents a sale transaction to update inventory and stats
type SaleTransaction struct {
	ProductID       uuid.UUID `json:"productId" validate:"required"`
	UnitsSold       int       `json:"unitsSold" validate:"required,min=1"`
	SaleAmount      float64   `json:"saleAmount" validate:"required,min=0"`
	TransactionDate time.Time `json:"transactionDate"`
}
