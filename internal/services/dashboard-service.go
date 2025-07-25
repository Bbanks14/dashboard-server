// File: internal/services/dashboard_service.go
package services

import (
	"time"

	"github.com/Bbanks14/dashboard-server/internal/data"
	"github.com/Bbanks14/dashboard-server/internal/domain"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/Bbanks14/dashboard-server/internal/services"
)

// NewDashboardService creates a new dashboard service
func NewDashboardService(db *data.Database) *services.DashboardService {
	return &services.DashboardService{
		db:                db,
		overallStatRepo:   repositories.NewOverallStatRepository(db),
		dailyStatRepo:     repositories.NewDailyStatRepository(db),
		monthlyStatRepo:   repositories.NewMonthlyStatRepository(db),
		productStatRepo:   repositories.NewProductStatRepository(db),
		geographyStatRepo: repositories.NewGeographyStatRepository(db),
		affiliateStatRepo: repositories.NewAffiliateStatRepository(db),
		transactionRepo:   repositories.NewTransactionRepository(db),
		customerRepo:      repositories.NewCustomerRepository(db),
	}
}

// GetDashboardData retrieves comprehensive dashboard data
func (s *DashboardService) GetDashboardData(year int) (*domain.DashboardData, error) {
	// Get overall stats for the year
	overallStats, err := s.overallStatRepo.GetByYear(year)
	if err != nil {
		return nil, err
	}

	// Get daily stats for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	dailyStats, err := s.dailyStatRepo.GetByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get monthly stats for the year
	monthlyStats, err := s.monthlyStatRepo.GetByYear(year)
	if err != nil {
		return nil, err
	}

	// Get top 10 products by sales
	topProducts, err := s.productStatRepo.GetTopProducts(year, 10)
	if err != nil {
		return nil, err
	}

	// Get geography stats
	geographyStats, err := s.geographyStatRepo.GetByYear(year)
	if err != nil {
		return nil, err
	}

	// Get recent transactions (last 20)
	recentTransactions, err := s.transactionRepo.GetRecent(20)
	if err != nil {
		return nil, err
	}

	// Get total counts
	totalCustomers, err := s.customerRepo.GetTotalCount()
	if err != nil {
		return nil, err
	}

	totalTransactions, err := s.transactionRepo.GetTotalCount()
	if err != nil {
		return nil, err
	}

	return &domain.DashboardData{
		OverallStats:       overallStats,
		DailyStats:         dailyStats,
		MonthlyStats:       monthlyStats,
		TopProducts:        topProducts,
		GeographyStats:     geographyStats,
		RecentTransactions: recentTransactions,
		TotalCustomers:     totalCustomers,
		TotalTransactions:  totalTransactions,
	}, nil
}

// GetOverviewData retrieves overview data (similar to dashboard but different focus)
func (s *DashboardService) GetOverviewData(year int) (*domain.DashboardData, error) {
	// This could be similar to GetDashboardData but with different metrics
	// For now, we'll use the same structure but could be customized
	return s.GetDashboardData(year)
}

// GetDailyStats retrieves daily statistics for a date range
func (s *DashboardService) GetDailyStats(startDate, endDate string) ([]domain.DailyStat, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	return s.dailyStatRepo.GetByDateRange(start, end)
}

// GetMonthlyStats retrieves monthly statistics for a year
func (s *DashboardService) GetMonthlyStats(year int) ([]domain.MonthlyStat, error) {
	return s.monthlyStatRepo.GetByYear(year)
}

// GetGeographyStats retrieves geography statistics for a year
func (s *DashboardService) GetGeographyStats(year int) ([]domain.GeographyStat, error) {
	return s.geographyStatRepo.GetByYear(year)
}

// GetPerformanceStats retrieves affiliate/user performance statistics
func (s *DashboardService) GetPerformanceStats(year int, month *int) ([]domain.AffiliateStat, error) {
	if month != nil {
		return s.affiliateStatRepo.GetByYearMonth(year, *month)
	}
	return s.affiliateStatRepo.GetByYear(year)
}

// GetBreakdownStats retrieves detailed breakdown statistics
func (s *DashboardService) GetBreakdownStats(metricType string, year int, month *int) (interface{}, error) {
	switch metricType {
	case "sales":
		if month != nil {
			return s.dailyStatRepo.GetByYearMonth(year, *month)
		}
		return s.monthlyStatRepo.GetByYear(year)
	case "products":
		return s.productStatRepo.GetByYear(year, month)
	case "customers":
		return s.customerRepo.GetStatsByYear(year)
	case "geography":
		return s.geographyStatRepo.GetByYear(year)
	default:
		return s.monthlyStatRepo.GetByYear(year)
	}
}

// GetAdminStats retrieves admin-specific statistics
func (s *DashboardService) GetAdminStats() (map[string]interface{}, error) {
	// Admin stats could include system-wide metrics
	totalUsers, err := s.customerRepo.GetTotalCount()
	if err != nil {
		return nil, err
	}

	totalTransactions, err := s.transactionRepo.GetTotalCount()
	if err != nil {
		return nil, err
	}

	// Get current year overall stats
	currentYear := time.Now().Year()
	overallStats, err := s.overallStatRepo.GetByYear(currentYear)
	if err != nil {
		return nil, err
	}

	// Get top performing affiliates
	topAffiliates, err := s.affiliateStatRepo.GetTopPerformers(currentYear, 10)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_users":        totalUsers,
		"total_transactions": totalTransactions,
		"overall_stats":      overallStats,
		"top_affiliates":     topAffiliates,
		"system_health":      s.getSystemHealth(),
	}, nil
}

// getSystemHealth returns basic system health metrics
func (s *DashboardService) getSystemHealth() map[string]interface{} {
	return map[string]interface{}{
		"database_status": "healthy",
		"last_updated":    time.Now(),
		"uptime":          "99.9%",
	}
}

// UpdateDashboardStats updates/refreshes dashboard statistics
func (s *DashboardService) UpdateDashboardStats() error {
	// This method would typically be called by a cron job or background process
	// to update all the statistical data

	currentTime := time.Now()
	currentYear := currentTime.Year()
	currentMonth := int(currentTime.Month())
	currentDate := currentTime.Truncate(24 * time.Hour)

	// Update daily stats
	if err := s.updateDailyStats(currentDate); err != nil {
		return err
	}

	// Update monthly stats
	if err := s.updateMonthlyStats(currentYear, currentMonth); err != nil {
		return err
	}

	// Update overall stats
	if err := s.updateOverallStats(currentYear); err != nil {
		return err
	}

	// Update product stats
	if err := s.updateProductStats(currentYear, currentMonth); err != nil {
		return err
	}

	// Update geography stats
	if err := s.updateGeographyStats(currentYear); err != nil {
		return err
	}

	// Update affiliate stats
	if err := s.updateAffiliateStats(currentYear, currentMonth); err != nil {
		return err
	}

	return nil
}

// Helper methods for updating stats
func (s *DashboardService) updateDailyStats(date time.Time) error {
	// Calculate daily statistics from transactions
	dailySales, dailyUnits, err := s.transactionRepo.GetDailyTotals(date)
	if err != nil {
		return err
	}

	return s.dailyStatRepo.Upsert(date, dailySales, dailyUnits)
}

func (s *DashboardService) updateMonthlyStats(year, month int) error {
	// Calculate monthly statistics from transactions
	monthlySales, monthlyUnits, err := s.transactionRepo.GetMonthlyTotals(year, month)
	if err != nil {
		return err
	}

	return s.monthlyStatRepo.Upsert(year, month, monthlySales, monthlyUnits)
}

func (s *DashboardService) updateOverallStats(year int) error {
	// Calculate overall statistics for the year
	totalCustomers, err := s.customerRepo.GetTotalCount()
	if err != nil {
		return err
	}

	yearlySales, yearlyUnits, err := s.transactionRepo.GetYearlyTotals(year)
	if err != nil {
		return err
	}

	return s.overallStatRepo.Upsert(year, totalCustomers, yearlySales, yearlyUnits)
}

func (s *DashboardService) updateProductStats(year, month int) error {
	// Calculate product statistics
	productStats, err := s.transactionRepo.GetProductStats(year, month)
	if err != nil {
		return err
	}

	return s.productStatRepo.UpsertBatch(productStats)
}

func (s *DashboardService) updateGeographyStats(year int) error {
	// Calculate geography statistics
	geographyStats, err := s.customerRepo.GetGeographyStats(year)
	if err != nil {
		return err
	}

	return s.geographyStatRepo.UpsertBatch(geographyStats)
}

func (s *DashboardService) updateAffiliateStats(year, month int) error {
	// Calculate affiliate statistics
	affiliateStats, err := s.transactionRepo.GetAffiliateStats(year, month)
	if err != nil {
		return err
	}

	return s.affiliateStatRepo.UpsertBatch(affiliateStats)
}
