package models

import (
	"time"

	"github.com/google/uuid"
)

// Unified Revenue History
type RevenueHistory struct {
	Date    time.Time `json:"date" db:"date"`
	Revenue float64   `json:"revenue" db:"revenue"`
}

// Conversion Metrics
type ConversionRate struct {
	Date           time.Time `json:"date" db:"date"`
	Opportunities  int64     `json:"opportunities" db:"opportunities"`
	Conversions    int64     `json:"conversions" db:"conversions"`
	ConversionRate float64   `json:"conversion_rate" db:"conversion_rate"`
}

// Forecast Information
type ForecastData struct {
	PeriodID       uuid.UUID `json:"period_id" db:"period_id"`
	ForecastAmount float64   `json:"forecast_amount" db:"forecast_amount"`
}

// Product Performance
type ProductSummary struct {
	ProductID   uuid.UUID `json:"product_id" db:"product_id"`
	ProductName string    `json:"product_name" db:"product_name"`
	UnitsSold   int64     `json:"units_sold" db:"units_sold"`
	TotalSales  float64   `json:"total_sales" db:"total_sales"`
}

// Regional Performance
type RegionSummary struct {
	RegionID      uuid.UUID `json:"region_id" db:"region_id"`
	RegionName    string    `json:"region_name" db:"region_name"`
	TotalQuantity int64     `json:"total_quantity" db:"total_quantity"`
	TotalSales    float64   `json:"total_sales" db:"total_sales"`
}

// Revenue Growth Models

// RevenueByDate represents revenue data for a specific date/period
type RevenueByDate struct {
	Date                time.Time `json:"date" db:"date"`
	Revenue             float64   `json:"revenue" db:"revenue"`
	TransactionCount    int64     `json:"transaction_count" db:"transaction_count"`
	AvgTransactionValue float64   `json:"avg_transaction_value" db:"avg_transaction_value"`
}

// PeriodComparison represents comparison between two time periods
type PeriodComparison struct {
	CurrentRevenue        float64 `json:"current_revenue" db:"current_revenue"`
	CurrentTransactions   int64   `json:"current_transactions" db:"current_transactions"`
	CurrentAvgValue       float64 `json:"current_avg_value" db:"current_avg_value"`
	PreviousRevenue       float64 `json:"previous_revenue" db:"previous_revenue"`
	PreviousTransactions  int64   `json:"previous_transactions" db:"previous_transactions"`
	PreviousAvgValue      float64 `json:"previous_avg_value" db:"previous_avg_value"`
	RevenueGrowthRate     float64 `json:"revenue_growth_rate" db:"revenue_growth_rate"`
	TransactionGrowthRate float64 `json:"transaction_growth_rate" db:"transaction_growth_rate"`
	AvgValueGrowthRate    float64 `json:"avg_value_growth_rate" db:"avg_value_growth_rate"`
}

// ProductRevenueSplit represents revenue breakdown by product
type ProductRevenueSplit struct {
	ProductID           uuid.UUID `json:"product_id" db:"product_id"`
	ProductName         string    `json:"product_name" db:"product_name"`
	Category            string    `json:"category" db:"category"`
	Revenue             float64   `json:"revenue" db:"revenue"`
	RevenuePercentage   float64   `json:"revenue_percentage" db:"revenue_percentage"`
	UnitsSold           int64     `json:"units_sold" db:"units_sold"`
	TransactionCount    int64     `json:"transaction_count" db:"transaction_count"`
	AvgTransactionValue float64   `json:"avg_transaction_value" db:"avg_transaction_value"`
}

// CustomerRevenueSplit represents revenue breakdown by individual customers
type CustomerRevenueSplit struct {
	CustomerID          uuid.UUID `json:"customer_id" db:"customer_id"`
	CustomerName        string    `json:"customer_name" db:"customer_name"`
	CustomerType        string    `json:"customer_type" db:"customer_type"`
	RegionName          string    `json:"region_name" db:"region_name"`
	Revenue             float64   `json:"revenue" db:"revenue"`
	RevenuePercentage   float64   `json:"revenue_percentage" db:"revenue_percentage"`
	TransactionCount    int64     `json:"transaction_count" db:"transaction_count"`
	AvgTransactionValue float64   `json:"avg_transaction_value" db:"avg_transaction_value"`
	FirstPurchaseDate   time.Time `json:"first_purchase_date" db:"first_purchase_date"`
	LastPurchaseDate    time.Time `json:"last_purchase_date" db:"last_purchase_date"`
}

// CustomerSegmentRevenue represents revenue breakdown by customer segments/types
type CustomerSegmentRevenue struct {
	Segment             string  `json:"segment" db:"segment"`
	Revenue             float64 `json:"revenue" db:"revenue"`
	RevenuePercentage   float64 `json:"revenue_percentage" db:"revenue_percentage"`
	CustomerCount       int64   `json:"customer_count" db:"customer_count"`
	TransactionCount    int64   `json:"transaction_count" db:"transaction_count"`
	AvgTransactionValue float64 `json:"avg_transaction_value" db:"avg_transaction_value"`
	RevenuePerCustomer  float64 `json:"revenue_per_customer" db:"revenue_per_customer"`
}

// RegionRevenueSplit represents revenue breakdown by regions
type RegionRevenueSplit struct {
	RegionID            uuid.UUID `json:"region_id" db:"region_id"`
	RegionName          string    `json:"region_name" db:"region_name"`
	Revenue             float64   `json:"revenue" db:"revenue"`
	RevenuePercentage   float64   `json:"revenue_percentage" db:"revenue_percentage"`
	CustomerCount       int64     `json:"customer_count" db:"customer_count"`
	TransactionCount    int64     `json:"transaction_count" db:"transaction_count"`
	AvgTransactionValue float64   `json:"avg_transaction_value" db:"avg_transaction_value"`
	RevenuePerCustomer  float64   `json:"revenue_per_customer" db:"revenue_per_customer"`
}

// Dashboard Aggregations

// DashboardMetrics aggregates key dashboard performance indicators
type DashboardMetrics struct {
	TotalRevenue     float64 `json:"total_revenue" db:"total_revenue"`
	TotalSales       int64   `json:"total_sales" db:"total_sales"`
	TargetSales      float64 `json:"target_sales" db:"target_sales"`
	ConversionRate   float64 `json:"conversion_rate" db:"conversion_rate"`
	TotalLeads       int64   `json:"total_leads" db:"total_leads"`
	TotalConversions int64   `json:"total_conversions" db:"total_conversions"`
}

// DashboardData represents the complete dashboard data structure
type DashboardData struct {
	Metrics         DashboardMetrics `json:"metrics"`
	RevenueHistory  []RevenueHistory `json:"revenue_history"`
	ConversionRates []ConversionRate `json:"conversion_rates"`
	Forecasts       []ForecastData   `json:"forecasts"`
	ProductSummary  []ProductSummary `json:"product_summary"`
	RegionSummary   []RegionSummary  `json:"region_summary"`
}

// RevenueGrowthData represents the complete revenue growth analysis
type RevenueGrowthData struct {
	RevenueByDate          []RevenueByDate          `json:"revenue_by_date"`
	PeriodComparison       PeriodComparison         `json:"period_comparison"`
	ProductRevenueSplit    []ProductRevenueSplit    `json:"product_revenue_split"`
	CustomerRevenueSplit   []CustomerRevenueSplit   `json:"customer_revenue_split"`
	CustomerSegmentRevenue []CustomerSegmentRevenue `json:"customer_segment_revenue"`
	RegionRevenueSplit     []RegionRevenueSplit     `json:"region_revenue_split"`
}

// RecentTransactions Data struct defines the total data structure for displaying all recent transactional activities
// including activity, order id, date, time, price and status
type RecentTransactions struct {
	ID      int       `json:"id" db:"id"`
	OrderID string    `json:"order_id" db:"order_id"`
	Date    time.Time `json:"date" db:"date"`
	Price   string    `json:"price" db:"price"`
	Status  string    `json:"status" db:"status"`
}

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	PictureURL string `json:"picture_url"
}

type Notification struct {
	ID        int       `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Message   string    `json:"message" db:"message"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SalesMetric struct {
	ID    int     `json:"id" db:"id"`
	UserID int     `json:"user_id" db:"user_id"`
	MetricName	string  `json:"metric_name" db:"metric_name"`	
	Value float64 `json:"value" db:"value"`
	Description string `json:"description" db:"description"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
