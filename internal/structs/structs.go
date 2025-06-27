package structs

import (
	"time"

	"github.com/google/uuid"
)

// DataAffiliateStruct defines data type expected as `json:""`
type DataAffiliateStruct struct {
	ID             string   `db:"id" json:"id"`
	UserId         string   `json:"userId"`
	AffiliateSales []string `json:"affiliateSales"`
}

// AffiliateStatSchema represents the schema for affiliate statistics in PostgreSQL.
type AffiliateStatSchema struct {
	ID             uuid.UUID `db:"id" json:"id"`
	UserID         uuid.UUID `db:"user_id" json:"userId"`
	AffiliateSales []string  `db:"affiliate_sales" json:"affiliateSales"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}

// DashboardStat represents analytics data for the dashboard
type DashboardStat struct {
	ID             int64     `db:"id" json:"id"`
	Date           time.Time `db:"date" json:"date"`
	TotalUsers     int       `db:"total_users" json:"total_users"`
	ActiveUsers    int       `db:"active_users" json:"active_users"`
	Revenue        float64   `db:"revenue" json:"revenue"`
	Transactions   int       `db:"transactions" json:"transactions"`
	ConversionRate float64   `db:"conversion_rate" json:"conversion_rate"`
}

// OverallStat defines the overall sales data type as expected in json
type OverallStat struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	TotalCustomers       int       `db:"customers" json:"totalCustomers"`
	YearlySalesTotal     int       `db:"yearly_sales_total" json:"yearlySalesTotal"`
	YearlyTotalSoldUnits int       `db:"yearly_total_sold_units" json:"yearlyTotalSoldUnits"`

	Year            int                 `db:"year" json:"year"`
	MonthlyData     []MonthlyDataStruct `db:"monthly_data" json:"monthlyData"`
	DailyData       []DailyDataStruct   `json:"dailyData"`
	SalesByCategory SalesByCategory     `db:"sales_by_category" json:"salesByCategory"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt string    `db:"updated_at" json:"updatedAt"`
	V         time.Time `db:"v" json:"__v"`
}

type MonthlyDataStruct struct {
	Month      string `db:"month" json:"month"`
	TotalSales int    `db:"total_sales" json:"totalSales"`
	TotalUnits int    `db:"total_units" json:"totalUnits"`
}

type DailyDataStruct struct {
	Date       string `db:"date" json:"date"`
	TotalSales int    `db:"total_sales" json:"totalSales"`
	TotalUnits int    `db:"total_units" json:"totalUnits"`
}

type SalesByCategory struct {
	Shoes       int `db:"shoes" json:"shoes"`
	Clothing    int `db:"clothing" json:"clothing"`
	Accessories int `db:"accessories" json:"accessories"`
	Misc        int `db:"misc" json:"misc"`
}
