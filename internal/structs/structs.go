package structs

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

// pgxProductRepoStruct defines a db struct using pgxpool
type pgxProductRepoStruct struct {
	db *pgxpool.Pool
}
