package models

import "time"

// DashboardStat represents analytics data for the dashboard
type DashboardStat struct {
	ID             int64     `json:"id"`
	Date           time.Time `json:"date"`
	TotalUsers     int       `json:"total_users"`
	ActiveUsers    int       `json:"active_users"`
	Revenue        float64   `json:"revenue"`
	Transactions   int       `json:"transactions"`
	ConversionRate float64   `json:"conversion_rate"`
}

// DashboardStats represents analytics data for the dashboard within a specific date range
type DashboardStats struct {
	TotalUsers         int64   // Total number of registered users
	NewUsers           int64   // Users created within the date range
	ActiveUsers        int64   // Users active within the date range
	AvgSessionDuration float64 // Average session duration in minutes
	PeakUsageTime      string  // Busiest time window in HH:MM-HH:MM format
}

// DashboardSummary provides aggregated user activity metrics
type DashboardSummary struct {
	TotalUsers      int64 // Total registered users across all time
	ActiveToday     int64 // Users active within the current day
	ActiveThisWeek  int64 // Users active within the current week
	ActiveThisMonth int64 // Users active within the current month
}
