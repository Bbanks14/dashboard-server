package services

import (
	"time"

	"github.com/Bbanks14/dashboard-server/internal/data/database"
	"github.com/Bbanks14/dashboard-server/internal/models"
)

// DashboardService interface defines the contract for dashboard services
type DashboardService interface {
	GetUser(startDate, endDate time.Time) ([]models.User, error)
	GetDashboardStats(startDate, endDate time.Time) (*models.DashboardStats, error)
	GetDashboardSummary() (*models.DashboardSummary, error)
}

// dashboardService implements the DashboardService interface
type dashboardService struct {
	DB *database.Database
}

// NewDashboardService creates a new dashboard service instance
func NewDashboardService(db *database.Database) DashboardService {
	return &dashboardService{
		DB: db,
	}
}

// GetUser retrieves user data for the given date range
func (s *dashboardService) GetUser(startDate, endDate time.Time) ([]models.User, error) {
	// Example implementation - you would replace this with actual database queries
	var users []models.User

	// Query for users created or active between the start and end dates
	if err := s.DB.Connection.
		Where("created_at BETWEEN ? AND ? OR last_active_at BETWEEN ? AND ?",
			startDate, endDate, startDate, endDate).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetDashboardStats retrieves analytics data for the dashboard based on date range
func (s *dashboardService) GetDashboardStats(startDate, endDate time.Time) (*models.DashboardStats, error) {
	// Initialize stats object
	stats := &models.DashboardStats{}

	// Count total users
	var totalUsers int64
	if err := s.DB.Connection.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers

	// Count new users in date range
	var newUsers int64
	if err := s.DB.Connection.Model(&models.User{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&newUsers).Error; err != nil {
		return nil, err
	}
	stats.NewUsers = newUsers

	// Count active users in date range
	var activeUsers int64
	if err := s.DB.Connection.Model(&models.User{}).
		Where("last_active_at BETWEEN ? AND ?", startDate, endDate).
		Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	stats.ActiveUsers = activeUsers

	// Calculate average session duration
	// This is a placeholder - implement according to your data model
	var avgSessionDuration float64
	rows, err := s.DB.Connection.Raw(`
		SELECT AVG(TIMESTAMPDIFF(MINUTE, session_start, session_end))
		FROM user_sessions
		WHERE session_start BETWEEN ? AND ?`,
		startDate, endDate).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&avgSessionDuration)
	}
	stats.avgSessionDuration = avgSessionDuration

	// Determine peak usage time - placeholder implementation
	// You would implement this according to your specific requirements
	starts.PeakUsageTime = "14:00-16:00"

	return stats, nil
}

// GetDashboardSummary retrieves a summary of current dashboard statistics
func (s *dashboardService) GetDashboardSummary() (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{
		TotalUsers:      0,
		ActiveToday:     0,
		ActiveThisWeek:  0,
		ActiveThisMonth: 0,
	}

	// Get total user count
	if err := s.DB.Connection.Model(&models.User{}).Count(&summary.TotalUsers).Error; err != nil {
		return nil, err
	}

	// Get users active today
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.DB.Connection.Model(&models.User{}).
		Where("last_active_at >= ?", today).
		Count(&summary.ActiveToday).Error; err != nil {
		return nil, err
	}

	// Get users active this week
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	if err := s.DB.Connection.Model(&models.User{}).
		Where("last_active_at >= ?", weekStart).
		Count(&summary.ActiveThisWeek).Error; err != nil {
		return nil, err
	}

	// Get users active this month
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	if err := s.DB.Connection.Model(&models.User{}).
		Where("last_active_at >= ?", monthStart).
		Count(&summary.ActiveThisMonth).Error; err != nil {
		return nil, err
	}

	return summary, nil
}
