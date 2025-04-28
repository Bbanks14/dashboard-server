// File: test/unit/services/dashboard_service_test.go
package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/dashboard-backend/internal/models"
	"github.com/yourusername/dashboard-backend/internal/services"
)

// MockDashboardStatRepository is a mock implementation of DashboardStatRepository
type MockDashboardStatRepository struct {
	mock.Mock
}

func (m *MockDashboardStatRepository) GetStats(startDate, endDate time.Time) ([]*models.DashboardStat, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]*models.DashboardStat), args.Error(1)
}

func (m *MockDashboardStatRepository) GetSummary() (*models.DashboardStat, error) {
	args := m.Called()
	return args.Get(0).(*models.DashboardStat), args.Error(1)
}

func (m *MockDashboardStatRepository) CreateStat(stat *models.DashboardStat) error {
	args := m.Called(stat)
	return args.Error(0)
}

func TestGetDashboardStats(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockDashboardStatRepository)

	// Create test data
	now := time.Now()
	startDate := now.AddDate(0, 0, -7)
	endDate := now

	expectedStats := []*models.DashboardStat{
		{
			ID:             1,
			Date:           now.AddDate(0, 0, -1),
			TotalUsers:     100,
			ActiveUsers:    80,
			Revenue:        1000.50,
			Transactions:   45,
			ConversionRate: 0.45,
		},
	}

	// Set up expectations
	mockRepo.On("GetStats", startDate, endDate).Return(expectedStats, nil)

	// Create service with mock repository
	service := services.NewDashboardService(mockRepo)

	// Call the method being tested
	stats, err := service.GetDashboardStats(startDate, endDate)

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	mockRepo.AssertExpectations(t)
}
