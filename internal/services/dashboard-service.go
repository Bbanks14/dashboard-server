package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/google/uuid"
)

type DashboardService interface {
	GetDashboardData(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (*models.DashboardData, error)
	GetDashboardMetrics(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (*models.DashboardMetrics, error)
	GetRevenueHistory(ctx context.Context, startDate, endDate string) ([]models.RevenueHistory, error)
	GetConversionRates(ctx context.Context, startDate, endDate string) ([]models.ConversionRate, error)
	GetForecastData(ctx context.Context, startDate, endDate string) ([]models.ForecastData, error)
	GetProductSummary(ctx context.Context, startDate, endDate string) ([]models.ProductSummary, error)
	GetRegionSummary(ctx context.Context, startDate, endDate string) ([]models.RegionSummary, error)
}

type DashboardRepositoryInterface interface {
	TotalSales(ctx context.Context, startDate, endDate string) (float64, int64, error)
	TargetSales(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (float64, error)
	RevenueHistory(ctx context.Context, startDate, endDate string) ([]models.RevenueHistory, error)
	SalesConversionRates(ctx context.Context, startDate, endDate string) ([]models.ConversionRate, error)
	ForecastData(ctx context.Context, startDate, endDate string) ([]models.ForecastData, error)
	ProductSummary(ctx context.Context, startDate, endDate string) ([]models.ProductSummary, error)
	RegionSummary(ctx context.Context, startDate, endDate string) ([]models.RegionSummary, error)
}

type dashboardService struct {
	repo DashboardRepositoryInterface
}

func NewDashboardService(repo DashboardRepositoryInterface) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetDashboardData(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (*models.DashboardData, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Fetch all data in parallel
	type result struct {
		metrics         *models.DashboardMetrics
		revenueHistory  []models.RevenueHistory
		conversionRates []models.ConversionRate
		forecasts       []models.ForecastData
		productSummary  []models.ProductSummary
		regionSummary   []models.RegionSummary
		err             error
	}

	ch := make(chan result, 1)
	go func() {
		res := result{}

		// Get metrics
		res.metrics, res.err = s.GetDashboardMetrics(ctx, startDate, endDate, regionID)
		if res.err != nil {
			ch <- res
			return
		}

		// Get other data concurrently
		res.revenueHistory, res.err = s.repo.RevenueHistory(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get revenue history: %w", res.err)
			ch <- res
			return
		}

		res.conversionRates, res.err = s.repo.SalesConversionRates(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get conversion rates: %w", res.err)
			ch <- res
			return
		}

		res.forecasts, res.err = s.repo.ForecastData(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get forecast data: %w", res.err)
			ch <- res
			return
		}

		res.productSummary, res.err = s.repo.ProductSummary(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get product summary: %w", res.err)
			ch <- res
			return
		}

		res.regionSummary, res.err = s.repo.RegionSummary(ctx, startDate, endDate)
		if res.err != nil {
			res.err = fmt.Errorf("failed to get region summary: %w", res.err)
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
		return &models.DashboardData{
			Metrics:         *res.metrics,
			RevenueHistory:  res.revenueHistory,
			ConversionRates: res.conversionRates,
			Forecasts:       res.forecasts,
			ProductSummary:  res.productSummary,
			RegionSummary:   res.regionSummary,
		}, nil
	}
}

func (s *dashboardService) GetDashboardMetrics(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (*models.DashboardMetrics, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get total sales
	totalRevenue, totalSales, err := s.repo.TotalSales(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sales: %w", err)
	}

	// Get target sales
	targetSales, err := s.repo.TargetSales(ctx, startDate, endDate, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target sales: %w", err)
	}

	// Get conversion rates
	conversionRates, err := s.repo.SalesConversionRates(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversion rates: %w", err)
	}

	// Calculate aggregated conversion metrics
	var totalLeads, totalConversions int64
	for _, rate := range conversionRates {
		totalLeads += rate.Opportunities
		totalConversions += rate.Conversions
	}

	overallConversionRate := 0.0
	if totalLeads > 0 {
		overallConversionRate = float64(totalConversions) / float64(totalLeads) * 100
	}

	return &models.DashboardMetrics{
		TotalRevenue:     totalRevenue,
		TotalSales:       totalSales,
		TargetSales:      targetSales,
		ConversionRate:   overallConversionRate,
		TotalLeads:       totalLeads,
		TotalConversions: totalConversions,
	}, nil
}

func (s *dashboardService) GetRevenueHistory(ctx context.Context, startDate, endDate string) ([]models.RevenueHistory, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}
	return s.repo.RevenueHistory(ctx, startDate, endDate)
}

func (s *dashboardService) GetConversionRates(ctx context.Context, startDate, endDate string) ([]models.ConversionRate, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}
	return s.repo.SalesConversionRates(ctx, startDate, endDate)
}

func (s *dashboardService) GetForecastData(ctx context.Context, startDate, endDate string) ([]models.ForecastData, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}
	return s.repo.ForecastData(ctx, startDate, endDate)
}

func (s *dashboardService) GetProductSummary(ctx context.Context, startDate, endDate string) ([]models.ProductSummary, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}
	return s.repo.ProductSummary(ctx, startDate, endDate)
}

func (s *dashboardService) GetRegionSummary(ctx context.Context, startDate, endDate string) ([]models.RegionSummary, error) {
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}
	return s.repo.RegionSummary(ctx, startDate, endDate)
}

// Helper methods

func (s *dashboardService) validateDateRange(startDate, endDate string) error {
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
