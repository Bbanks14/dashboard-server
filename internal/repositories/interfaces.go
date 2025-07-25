// File: internal/repositories/interfaces.go
package repositories

import (
	"time"

	"github.com/Bbanks14/dashboard-server/internal/domain"
	"github.com/google/uuid"
)

// OverallStatRepository handles overall statistics operations
type OverallStatRepository interface {
	GetByYear(year int) (*domain.OverallStat, error)
	Upsert(year int, totalCustomers int, yearlySales float64, yearlyUnits int) error
	Create(stat *domain.OverallStat) error
	Update(stat *domain.OverallStat) error
	Delete(id uuid.UUID) error
}

// DailyStatRepository handles daily statistics operations
type DailyStatRepository interface {
	GetByDateRange(startDate, endDate time.Time) ([]domain.DailyStat, error)
	GetByDate(date time.Time) (*domain.DailyStat, error)
	GetByYearMonth(year, month int) ([]domain.DailyStat, error)
	Upsert(date time.Time, totalSales float64, totalUnits int) error
	Create(stat *domain.DailyStat) error
	Update(stat *domain.DailyStat) error
	Delete(id uuid.UUID) error
}

// MonthlyStatRepository handles monthly statistics operations
type MonthlyStatRepository interface {
	GetByYear(year int) ([]domain.MonthlyStat, error)
	GetByYearMonth(year, month int) (*domain.MonthlyStat, error)
	Upsert(year, month int, totalSales float64, totalUnits int) error
	Create(stat *domain.MonthlyStat) error
	Update(stat *domain.MonthlyStat) error
	Delete(id uuid.UUID) error
}

// ProductStatRepository handles product statistics operations
type ProductStatRepository interface {
	GetByYear(year int, month *int) ([]domain.ProductStat, error)
	GetTopProducts(year int, limit int) ([]domain.ProductStat, error)
	GetByProductID(productID uuid.UUID, year int, month *int) ([]domain.ProductStat, error)
	UpsertBatch(stats []domain.ProductStat) error
	Create(stat *domain.ProductStat) error
	Update(stat *domain.ProductStat) error
	Delete(id uuid.UUID) error
}

// GeographyStatRepository handles geography statistics operations
type GeographyStatRepository interface {
	GetByYear(year int) ([]domain.GeographyStat, error)
	GetByCountry(country string, year int) (*domain.GeographyStat, error)
	UpsertBatch(stats []domain.GeographyStat) error
	Create(stat *domain.GeographyStat) error
	Update(stat *domain.GeographyStat) error
	Delete(id uuid.UUID) error
}

// AffiliateStatRepository handles affiliate statistics operations
type AffiliateStatRepository interface {
	GetByYear(year int) ([]domain.AffiliateStat, error)
	GetByYearMonth(year, month int) ([]domain.AffiliateStat, error)
	GetByUserID(userID uuid.UUID, year int, month *int) ([]domain.AffiliateStat, error)
	GetTopPerformers(year int, limit int) ([]domain.AffiliateStat, error)
	UpsertBatch(stats []domain.AffiliateStat) error
	Create(stat *domain.AffiliateStat) error
	Update(stat *domain.AffiliateStat) error
	Delete(id uuid.UUID) error
}

// TransactionRepository handles transaction operations
type TransactionRepository interface {
	GetAll(limit, offset int) ([]domain.Transaction, error)
	GetByID(id uuid.UUID) (*domain.Transaction, error)
	GetByUserID(userID uuid.UUID, limit, offset int) ([]domain.Transaction, error)
	GetByCustomerID(customerID uuid.UUID, limit, offset int) ([]domain.Transaction, error)
	GetByDateRange(startDate, endDate time.Time, limit, offset int) ([]domain.Transaction, error)
	GetRecent(limit int) ([]domain.Transaction, error)
	GetTotalCount() (int, error)

	// Statistics methods
	GetDailyTotals(date time.Time) (totalSales float64, totalUnits int, err error)
	GetMonthlyTotals(year, month int) (totalSales float64, totalUnits int, err error)
	GetYearlyTotals(year int) (totalSales float64, totalUnits int, err error)
	GetProductStats(year, month int) ([]domain.ProductStat, error)
	GetAffiliateStats(year, month int) ([]domain.AffiliateStat, error)

	// CRUD operations
	Create(transaction *domain.Transaction) error
	Update(transaction *domain.Transaction) error
	Delete(id uuid.UUID) error
}

// CustomerRepository handles customer operations
type CustomerRepository interface {
	GetAll(limit, offset int) ([]domain.Customer, error)
	GetByID(id uuid.UUID) (*domain.Customer, error)
	GetByEmail(email string) (*domain.Customer, error)
	GetByCountry(country string, limit, offset int) ([]domain.Customer, error)
	GetTotalCount() (int, error)
	GetCountries() ([]string, error)
	GetGeographyStats(year int) ([]domain.GeographyStat, error)
	GetStatsByYear(year int) (map[string]interface{}, error)

	// Search and filter
	Search(query string, limit, offset int) ([]domain.Customer, error)
	Filter(filters map[string]interface{}, limit, offset int) ([]domain.Customer, error)

	// CRUD operations
	Create(customer *domain.Customer) error
	Update(customer *domain.Customer) error
	Delete(id uuid.UUID) error
}

// ProductRepository handles product operations
type ProductRepository interface {
	GetAll(limit, offset int) ([]domain.Product, error)
	GetByID(id uuid.UUID) (*domain.Product, error)
	GetByCategory(category string, limit, offset int) ([]domain.Product, error)
	GetCategories() ([]string, error)
	GetTotalCount() (int, error)

	// Search and filter
	Search(query string, limit, offset int) ([]domain.Product, error)
	Filter(filters map[string]interface{}, limit, offset int) ([]domain.Product, error)

	// CRUD operations
	Create(product *domain.Product) error
	Update(product *domain.Product) error
	Delete(id uuid.UUID) error
}

// UserRepository handles user operations
type UserRepository interface {
	GetAll(limit, offset int) ([]domain.User, error)
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByRole(role string, limit, offset int) ([]domain.User, error)
	GetTotalCount() (int, error)

	// CRUD operations
	Create(user *domain.User) error
	Update(user *domain.User) error
	Delete(id uuid.UUID) error
}
