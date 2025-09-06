package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/models"
)

// RevenueGrowthService defines the interface for revenue growth operations
type RevenueGrowthService interface {
	GetRevenueGrowthData(ctx context.Context, startDate, endDate string) (*models.RevenueGrowthData, error)
	GetRevenueByDate(ctx context.Context, startDate, endDate string, granularity string) ([]models.RevenueByDate, error)
	GetPeriodSalesComparison(ctx context.Context, currentStart, currentEnd, previousStart, previousEnd string) (*models.PeriodComparison, error)
	GetProductRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.ProductRevenueSplit, error)
	GetCustomerRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.CustomerRevenueSplit, error)
	GetCustomerSegmentRevenue(ctx context.Context, startDate, endDate string) ([]models.CustomerSegmentRevenue, error)
	GetRegionRevenueSplit(ctx context.Context, startDate, endDate string) ([]models.RegionRevenueSplit, error)
}

// RevenueGrowthRepositoryInterface defines what the service expects from the repository
type RevenueGrowthRepositoryInterface interface {
	RevenueByDate(ctx context.Context, startDate, endDate, granularity string) ([]models.RevenueByDate, error)
	PeriodSalesComparison(ctx context.Context, currentStart, currentEnd, previousStart, previousEnd string) (models.PeriodComparison, error)
	ProductRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.ProductRevenueSplit, error)
	CustomerRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.CustomerRevenueSplit, error)
	CustomerSegmentRevenue(ctx context.Context, startDate, endDate string) ([]models.CustomerSegmentRevenue, error)
	RegionRevenueSplit(ctx context.Context, startDate, endDate string) ([]models.RegionRevenueSplit, error)
}

// revenueGrowthService is the implementation of RevenueGrowthService
type revenueGrowthService struct {
	repo RevenueGrowthRepositoryInterface
}

// NewRevenueGrowthService creates a new instance of RevenueGrowthService
func NewRevenueGrowthService(repo RevenueGrowthRepositoryInterface) RevenueGrowthService {
	return &revenueGrowthService{repo: repo}
}

// GetRevenueGrowthData retrieves all revenue growth data in a single call
func (s *revenueGrowthService) GetRevenueGrowthData(ctx context.Context, startDate, endDate string) (*models.RevenueGrowthData, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Fetch all data in parallel
	type result struct {
		byDate           []models.RevenueByDate
		comparison       models.PeriodComparison
		productSplit     []models.ProductRevenueSplit
		customerSplit    []models.CustomerRevenueSplit
		customerSegments []models.CustomerSegmentRevenue
		regionSplit      []models.RegionRevenueSplit
		err              error
	}

	ch := make(chan result, 1)
	go func() {
		res := result{}

		// Get revenue by date (daily granularity)
		res.byDate, res.err = s.repo.RevenueByDate(ctx, startDate, endDate, "daily")
		if res.err != nil {
			ch <- res
			return
		}

		// Calculate comparison periods based on date range
		currentStart, currentEnd, previousStart, previousEnd, err := s.calculateComparisonPeriods(startDate, endDate)
		if err != nil {
			res.err = fmt.Errorf("failed to calculate comparison periods: %w", err)
			ch <- res
			return
		}

		// Get period comparison
		res.comparison, res.err = s.repo.PeriodSalesComparison(ctx, currentStart, currentEnd, previousStart, previousEnd)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get period comparison: %w", res.err)
			ch <- res
			return
		}

		// Get product revenue split (top 10 products)
		res.productSplit, res.err = s.repo.ProductRevenueSplit(ctx, startDate, endDate, 10)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get product revenue split: %w", res.err)
			ch <- res
			return
		}

		// Get customer revenue split (top 10 customers)
		res.customerSplit, res.err = s.repo.CustomerRevenueSplit(ctx, startDate, endDate, 10)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get customer revenue split: %w", res.err)
			ch <- res
			return
		}

		// Get customer segment revenue
		res.customerSegments, res.err = s.repo.CustomerSegmentRevenue(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get customer segment revenue: %w", res.err)
			ch <- res
			return
		}

		// Get region revenue split
		res.regionSplit, res.err = s.repo.RegionRevenueSplit(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get region revenue split: %w", res.err)
			ch <- res
			return
		}

		ch <- res
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}

		return &models.RevenueGrowthData{
			RevenueByDate:          res.byDate,
			PeriodComparison:       res.comparison,
			ProductRevenueSplit:    res.productSplit,
			CustomerRevenueSplit:   res.customerSplit,
			CustomerSegmentRevenue: res.customerSegments,
			RegionRevenueSplit:     res.regionSplit,
		}, nil
	}
}

// GetRevenueByDate retrieves revenue data grouped by date/period
func (s *revenueGrowthService) GetRevenueByDate(ctx context.Context, startDate, endDate string, granularity string) ([]models.RevenueByDate, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	if granularity == "" {
		granularity = "daily" // Default granularity
	}

	return s.repo.RevenueByDate(ctx, startDate, endDate, granularity)
}

// GetPeriodSalesComparison retrieves sales comparison between periods
func (s *revenueGrowthService) GetPeriodSalesComparison(ctx context.Context, currentStart, currentEnd, previousStart, previousEnd string) (*models.PeriodComparison, error) {
	// Validate all date ranges
	for _, dates := range [][]string{{currentStart, currentEnd}, {previousStart, previousEnd}} {
		if err := s.validateDateRange(dates[0], dates[1]); err != nil {
			return nil, err
		}
	}

	comparison, err := s.repo.PeriodSalesComparison(ctx, currentStart, currentEnd, previousStart, previousEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get period comparison: %w", err)
	}

	return &comparison, nil
}

// GetProductRevenueSplit retrieves revenue breakdown by products
func (s *revenueGrowthService) GetProductRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.ProductRevenueSplit, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}

	return s.repo.ProductRevenueSplit(ctx, startDate, endDate, limit)
}

// GetCustomerRevenueSplit retrieves revenue breakdown by customers
func (s *revenueGrowthService) GetCustomerRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.CustomerRevenueSplit, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}

	return s.repo.CustomerRevenueSplit(ctx, startDate, endDate, limit)
}

// GetCustomerSegmentRevenue retrieves revenue breakdown by customer segments
func (s *revenueGrowthService) GetCustomerSegmentRevenue(ctx context.Context, startDate, endDate string) ([]models.CustomerSegmentRevenue, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	return s.repo.CustomerSegmentRevenue(ctx, startDate, endDate)
}

// GetRegionRevenueSplit retrieves revenue breakdown by regions
func (s *revenueGrowthService) GetRegionRevenueSplit(ctx context.Context, startDate, endDate string) ([]models.RegionRevenueSplit, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	return s.repo.RegionRevenueSplit(ctx, startDate, endDate)
}

// Helper methods

// validateDateRange validates the provided date range
func (s *revenueGrowthService) validateDateRange(startDate, endDate string) error {
	if startDate == "" || endDate == "" {
		return fmt.Errorf("start date and end date are required")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	if end.Before(start) {
		return fmt.Errorf("end date must be after start date")
	}

	// Optional: Add maximum date range validation
	if end.Sub(start) > 365*24*time.Hour {
		return fmt.Errorf("date range cannot exceed 365 days")
	}

	return nil
}

// calculateComparisonPeriods determines previous period based on current date range
func (s *revenueGrowthService) calculateComparisonPeriods(startDate, endDate string) (string, string, string, string, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return "", "", "", "", err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return "", "", "", "", err
	}

	// Calculate duration of current period
	duration := end.Sub(start)

	// Calculate previous period
	prevStart := start.Add(-duration)
	prevEnd := end.Add(-duration)

	return startDate, endDate,
		prevStart.Format("2006-01-02"),
		prevEnd.Format("2006-01-02"), nil
}

func SearchApp(query string, userID uint) ([]models.SalesMetric, error) {
	return repositories.SearchSalesMetrics(userID, query)
}

func PullSalesData(userID uint) error {
	return repositories.PullSalesData(userID)
}
