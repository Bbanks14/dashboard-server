// File: internal/models/dashboard_stat.go
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

// DashboardStatRepository interface for dashboard stats operations
type DashboardStatRepository interface {
	GetStats(startDate, endDate time.Time) ([]*DashboardStat, error)
	GetSummary() (*DashboardStat, error)
	CreateStat(stat *DashboardStat) error
}
