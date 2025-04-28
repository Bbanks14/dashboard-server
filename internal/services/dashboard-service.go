// File: internal/services/dashboard_service.go
package services

import (
	"time"

	"github.com/yourusername/dashboard-backend/internal/data"
	"github.com/yourusername/dashboard-backend/internal/models"
)

// DashboardService handles business logic for dashboard analytics
type DashboardService struct {
	db       *data.Database
	statRepo models.DashboardStatRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(db *data.Database) *DashboardService {
	return &DashboardService{
		db:       db,
		statRepo: data.NewDashboardStatRepository(db),
	}
}

// GetDashboardStats retrieves analytics data for the dashboard
func (s *DashboardService) GetDashboardStats(startDate, endDate time.Time) ([]*models.DashboardStat, error) {
	return s.statRepo.GetStats(startDate, endDate)
}

// GetDashboardSummary retrieves a summary of the current dashboard stats
func (s *DashboardService) GetDashboardSummary() (*models.DashboardStat, error) {
	return s.statRepo.GetSummary()
}
