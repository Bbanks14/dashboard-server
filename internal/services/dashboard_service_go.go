package services

import (
	"context"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/data/database"
	"github.com/Bbanks14/dashboard-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DashboardService interface defines the contract for dashboard services
type DashboardService interface {
	GetUser(ctx context.Context, startDate, endDate time.Time) ([]models.User, error)
	GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (*models.DashboardStats, error)
	GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error)
}

// dashboardService implements the DashboardService interface
type dashboardService struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewDashboardService creates a new dashboard service instance
func NewDashboardService(client *mongo.Client, databaseName string) DashboardService {
	return &dashboardService{
		client:   client,
		database: client.Database(databaseName),
	}
}

// GetUser retrieves user data for the given date range
func (s *dashboardService) GetUser(startDate, endDate time.Time) ([]models.User, error) {
	// Example implementation - you would replace this with actual database queries
	var users []models.User

	// Query for users created or active between the start and end dates
	filter := bson.M{
		"$or": []bson.M{
			{"created_at": bson.M{"$gte": startDate, "$lte": endDate}},
			{"last_active_at": bson.M{"$gte": startDate, "$lte": endDate}},
		},
	}

	// Execute the query
	cursor, err := s.database.Collection("users").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetDashboardStats retrieves analytics data for the dashboard based on date range
func (s *dashboardService) GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (*models.DashboardStats, error) {
	// Initialize stats object
	stats := &models.DashboardStats{}

	// Count total users
	totalUsers, err := s.database.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers

	// Count new users in date range
	newUsers, err := s.database.Collection("users").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": startDate, "$lte": endDate},
	})
	if err != nil {
		return nil, err
	}
	stats.NewUsers = newUsers

	// Count active users in date range
	activeUsers, err := s.database.Collection("users").CountDocuments(ctx, bson.M{
		"last_active_at": bson.M{"$gte": startDate, "$lte": endDate},
	})
	if err != nil {
		return nil, err
	}
	stats.ActiveUsers = activeUsers

	// Calculate average session duration
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"session_start": bson.M{"$gte": startDate, "$lte": endDate},
		}}},
		{{Key: "$project", Value: bson.M{
			"duration_minutes": bson.M{
				"$divide": []interface{}{
					bson.M{"$subtract": []interface{}{"$session_end", "$session_start"}},
					60000, // Convert milliseconds to minutes
				},
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"avgDuration": bson.M{"$avg": "$duration_minutes"},
		}}},
	}

	cursor, err := s.database.Collection("user_sessions").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	var avgSessionDuration float64
	if len(result) > 0 && result[0]["avgDuration"] != nil {
		avgSessionDuration = result[0]["avgDuration"].(float64)
	}
	stats.AvgSessionDuration = avgSessionDuration

	// Determine peak usage time - placeholder implementation
	stats.PeakUsageTime = "14:00-16:00" // Fixed variable name

	return stats, nil
}

// GetDashboardSummary retrieves a summary of current dashboard statistics
func (s *dashboardService) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{
		TotalUsers:      0,
		ActiveToday:     0,
		ActiveThisWeek:  0,
		ActiveThisMonth: 0,
	}

	// Get total user count
	totalUsers, err := s.database.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	summary.TotalUsers = totalUsers

	// Get users active today
	today := time.Now().Truncate(24 * time.Hour)
	activeToday, err := s.database.Collection("users").CountDocuments(ctx, bson.M{
		"last_active_at": bson.M{"$gte": today},
	})
	if err != nil {
		return nil, err
	}
	summary.ActiveToday = activeToday

	// Get users active this week
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	activeThisWeek, err := s.database.Collection("users").CountDocuments(ctx, bson.M{
		"last_active_at": bson.M{"$gte": weekStart},
	})
	if err != nil {
		return nil, err
	}
	summary.ActiveThisWeek = activeThisWeek

	// Get users active this month
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	activeThisMonth, err := s.database.Collection("users").CountDocuments(ctx, bson.M{
		"last_active_at": bson.M{"$gte": monthStart},
	})
	if err != nil {
		return nil, err
	}
	summary.ActiveThisMonth = activeThisMonth

	return summary, nil
}
