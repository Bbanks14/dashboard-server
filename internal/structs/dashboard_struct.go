package structs

import (
	"github.com/Bbanks14/dashboard-server/internal/data"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
)

// DashboardService struct defines and handles business logic for dashboard analytics
type DashboardService struct {
	db                *data.Database
	overallStatRepo   repositories.OverallStatRepository
	dailyStatRepo     repositories.DailyStatRepository
	monthlyStatRepo   repositories.MonthlyStatRepository
	productStatRepo   repositories.ProductStatRepository
	geographyStatRepo repositories.GeographyStatRepository
	affiliateStatRepo repositories.AffiliateStatRepository
	transctionRepo    repositories.TransactionRepository
	customerRepo      repositories.CustomerRepository
}
