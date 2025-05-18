package structs

import (
	"time"
)

// DataAffiliateStruct defines data type expected as `json:""`
type DataAffiliateStruct struct {
	Id             string   `json:"_id"`
	UserId         string   `json:"userId"`
	AffiliateSales []string `json:"affiliateSales"`
}

// DashboardStat represents analytics data for the dashboard
type DashboardStatStruct struct {
	ID             int64     `json:"id"`
	Date           time.Time `json:"date"`
	TotalUsers     int       `json:"total_users"`
	ActiveUsers    int       `json:"active_users"`
	Revenue        float64   `json:"revenue"`
	Transactions   int       `json:"transactions"`
	ConversionRate float64   `json:"conversion_rate"`
}

// OverallStat defines the overall sales data type as expected in json
type OverallStatStruct struct {
	TotalCustomers       int `json:"totalCustomers"`
	YearlySalesTotal     int `json:"yearlySalesTotal"`
	YearlyTotalSoldUnits int `json:"yearlyTotalSoldUnits"`

	Year            int                 `json:"year"`
	MonthlyData     []MonthlyDataStruct `json:"monthlyData"`
	DailyData       []DailyDataStruct   `json:"dailyData"`
	SalesByCategory SalesByCategory     `json:"salesByCategory"`

	ID        string `json:"_id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	V         int    `json:"__v"`
}

type MonthlyDataStruct struct {
	Month      string `json:"month"`
	TotalSales int    `json:"totalSales"`
	TotalUnits int    `json:"totalUnits"`
	ID         string `json:"_id"`
}

type DailyDataStruct struct {
	Date       string `json:"date"`
	TotalSales int    `json:"totalSales"`
	TotalUnits int    `json:"totalUnits"`
}

type SalesByCategory struct {
	Shoes       int `json:"shoes"`
	Clothing    int `json:"clothing"`
	Accessories int `json:"accessories"`
	Misc        int `json:"misc"`
}
