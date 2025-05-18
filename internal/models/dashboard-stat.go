package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Email        string             `bson:"email" json:"email"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	LastActiveAt time.Time          `bson:"last_active_at" json:"last_active_at"`
}

// DashboardStat represents analytics data for the dashboard
type DashboardStat struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Date           time.Time          `bson:"date" json:"date"`
	TotalUsers     int                `bson:"total_users" json:"total_users"`
	ActiveUsers    int                `bson:"active_users" json:"active_users"`
	Revenue        float64            `bson:"revenue" json:"revenue"`
	Transactions   int                `bson:"transactions" json:"transactions"`
	ConversionRate float64            `bson:"conversion_rate" json:"conversion_rate"`
}

// DashboardStats represents analytics data for the dashboard within a specific date range
type DashboardStats struct {
	TotalUsers         int64   `bson:"total_users" json:"total_users"`
	NewUsers           int64   `bson:"new_users" json:"new_users"`
	ActiveUsers        int64   `bson:"active_users" json:"active_users"`
	AvgSessionDuration float64 `bson:"avg_sesion_duration" json:"avg_sesion_duration"`
	PeakUsageTime      string  `bson:"peak_usage_time" json:"peak_usage_time"`
}

// DashboardSummary provides aggregated user activity metrics
type DashboardSummary struct {
	TotalUsers      int64 `bson:"total_users" json:"total_users"`
	ActiveToday     int64 `bson:"active_today" json:"active_today"`
	ActiveThisWeek  int64 `bson:"active_this_week" json:"active_this_week"`
	ActiveThisMonth int64 `bson:"active_this_month" json:"active_this_month"`
}
