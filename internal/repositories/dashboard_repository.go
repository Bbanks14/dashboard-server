
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardRepository handles data access for dashboard metrics
type DashboardRepository struct {
	db *pgxpool.Pool
}

// NewDashboardRepository creates a new DashboardRepository
func NewDashboardRepository(db *pgxpool.Pool) (*DashboardRepository, error) {
	repo := &DashboardRepository{db: db}
	// Create necessary tables if they don't exist
	if err := repo.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables required for the dashboard: %w", err)
	}
	return repo, nil
}


// ===== DASHBOARD METHODS =====

// TotalSales retrieves the total sales amount and count for a date range
func (dr *DashboardRepository) TotalSales(ctx context.Context, startDate, endDate string) (float64, int64, error) {
	query := `
		SELECT COALESCE(SUM(total_amount), 0) as total_revenue, COUNT(*) as total_sales
		FROM sales
		WHERE sale_date BETWEEN $1 AND $2;
	`
	var totalRevenue float64
	var totalSales int64
	err := dr.db.QueryRow(ctx, query, startDate, endDate).Scan(&totalRevenue, &totalSales)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to retrieve total sales: %w", err)
	}
	return totalRevenue, totalSales, nil
}

// TargetSales retrieves the target sales amount for a date range and optional region
func (dr *DashboardRepository) TargetSales(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(target_amount), 0)
		FROM sales_targets st
		JOIN periods p ON st.period_id = p.period_id
		WHERE p.period_start <= $2 AND p.period_end >= $1
	`
	args := []interface{}{startDate, endDate}

	if regionID != nil {
		query += " AND st.region_id = $3"
		args = append(args, *regionID)
	}

	var targetAmount float64
	err := dr.db.QueryRow(ctx, query, args...).Scan(&targetAmount)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve target sales: %w", err)
	}
	return targetAmount, nil
}

// RevenueHistory retrieves revenue data grouped by date for a date range
func (dr *DashboardRepository) RevenueHistory(ctx context.Context, startDate, endDate string) ([]models.RevenueHistory, error) {
	query := `
		SELECT sale_date::DATE as date, COALESCE(SUM(total_amount), 0) as revenue
		FROM sales
		WHERE sale_date BETWEEN $1 AND $2
		GROUP BY sale_date::DATE
		ORDER BY sale_date::DATE;
	`
	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve revenue history: %w", err)
	}
	defer rows.Close()

	var history []models.RevenueHistory
	for rows.Next() {
		var rh models.RevenueHistory
		var date time.Time
		if err := rows.Scan(&date, &rh.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan revenue history: %w", err)
		}
		rh.Date = date
		history = append(history, rh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over revenue history: %w", err)
	}

	return history, nil
}

// SalesConversionRates retrieves conversion rates based on leads for a date range
func (dr *DashboardRepository) SalesConversionRates(ctx context.Context, startDate, endDate string) ([]models.ConversionRate, error) {
	query := `
		SELECT created_at::DATE as date, 
			COUNT(*) as opportunities, 
			COUNT(*) FILTER (WHERE status = 'converted') as conversions
		FROM leads
				COUNT(*) as transaction_count,
				COALESCE(AVG(COALESCE(total_amount, unit_price * quantity)), 0) as avg_transaction_value
			FROM sales
			WHERE sale_date BETWEEN $3 AND $4
		)
		SELECT
			cp.revenue as current_revenue,
			cp.transaction_count as current_transactions,
			cp.avg_transaction_value as current_avg_value,
			pp.revenue as previous_revenue,
			pp.transaction_count as previous_transactions,
			pp.avg_transaction_value as previous_avg_value
		FROM current_period cp, previous_period pp;
	`

	var pc models.PeriodComparison
	err := dr.db.QueryRow(ctx, query, currentPeriodStart, currentPeriodEnd, previousPeriodStart, previousPeriodEnd).Scan(
		&pc.CurrentRevenue, &pc.CurrentTransactions, &pc.CurrentAvgValue,
		&pc.PreviousRevenue, &pc.PreviousTransactions, &pc.PreviousAvgValue,
	)
	if err != nil {
		return pc, fmt.Errorf("failed to retrieve period sales comparison: %w", err)
	}

	// Calculate growth percentages
	if pc.PreviousRevenue > 0 {
		pc.RevenueGrowthRate = ((pc.CurrentRevenue - pc.PreviousRevenue) / pc.PreviousRevenue) * 100
	}
	if pc.PreviousTransactions > 0 {
		pc.TransactionGrowthRate = ((float64(pc.CurrentTransactions) - float64(pc.PreviousTransactions)) / float64(pc.PreviousTransactions)) * 100
	}
	if pc.PreviousAvgValue > 0 {
		pc.AvgValueGrowthRate = ((pc.CurrentAvgValue - pc.PreviousAvgValue) / pc.PreviousAvgValue) * 100
	}

	return pc, nil
}

func (dr *DashboardRepository) ProductRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.ProductRevenueSplit, error) {
	query := `
		SELECT
			p.product_id,
			p.name as product_name,
			p.category,
			COALESCE(SUM(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as revenue,
			COALESCE(SUM(s.quantity), 0) as units_sold,
			COUNT(s.sale_id) as transaction_count,
			COALESCE(AVG(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as avg_transaction_value
		FROM products p
		LEFT JOIN sales s ON p.product_id = s.product_id
			AND s.sale_date BETWEEN $1 AND $2
		GROUP BY p.product_id, p.name, p.category
		ORDER BY revenue DESC
		LIMIT $3;
	`

	rows, err := dr.db.Query(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve product revenue split: %w", err)
	}
	defer rows.Close()

	var splits []models.ProductRevenueSplit
	var totalRevenue float64
	for rows.Next() {
		var prs models.ProductRevenueSplit
		if err := rows.Scan(&prs.ProductID, &prs.ProductName, &prs.Category,
			&prs.Revenue, &prs.UnitsSold, &prs.TransactionCount, &prs.AvgTransactionValue); err != nil {
			return nil, fmt.Errorf("failed to scan product revenue split: %w", err)
		}
		totalRevenue += prs.Revenue
		splits = append(splits, prs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over product revenue split: %w", err)
	}

	// Calculate percentages
	for i := range splits {
		if totalRevenue > 0 {
			splits[i].RevenuePercentage = (splits[i].Revenue / totalRevenue) * 100
		}
	}

	return splits, nil
}

func (dr *DashboardRepository) CustomerRevenueSplit(ctx context.Context, startDate, endDate string, limit int) ([]models.CustomerRevenueSplit, error) {
	query := `
		SELECT
			c.customer_id,
			c.name as customer_name,
			c.customer_type,
			r.name as region_name,
			COALESCE(SUM(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as revenue,
			COUNT(s.sale_id) as transaction_count,
			COALESCE(AVG(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as avg_transaction_value,
			MIN(s.sale_date) as first_purchase_date,
			MAX(s.sale_date) as last_purchase_date
		FROM customers c
		LEFT JOIN regions r ON c.region_id = r.region_id
		LEFT JOIN sales s ON c.customer_id = s.customer_id
			AND s.sale_date BETWEEN $1 AND $2
		GROUP BY c.customer_id, c.name, c.customer_type, r.name
		HAVING COALESCE(SUM(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) > 0
		ORDER BY revenue DESC
		LIMIT $3;
	`

	rows, err := dr.db.Query(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer revenue split: %w", err)
	}
	defer rows.Close()

	var splits []models.CustomerRevenueSplit
	var totalRevenue float64
	for rows.Next() {
		var crs models.CustomerRevenueSplit
		var firstPurchase, lastPurchase *time.Time
		if err := rows.Scan(&crs.CustomerID, &crs.CustomerName, &crs.CustomerType,
			&crs.RegionName, &crs.Revenue, &crs.TransactionCount, &crs.AvgTransactionValue,
			&firstPurchase, &lastPurchase); err != nil {
			return nil, fmt.Errorf("failed to scan customer revenue splits: %w", err)
		}
		if firstPurchase != nil {
			crs.FirstPurchaseDate = *firstPurchase
		}
		if lastPurchase != nil {
			crs.LastPurchaseDate = *lastPurchase
		}
		totalRevenue += crs.Revenue
		splits = append(splits, crs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over customer revenue split: %w", err)
	}

	// Calculate percentages
	for i := range splits {
		if totalRevenue > 0 {
			splits[i].RevenuePercentage = (splits[i].Revenue / totalRevenue) * 100
		}
	}

	return splits, nil
}

func (dr *DashboardRepository) CustomerSegmentRevenue(ctx context.Context, startDate, endDate string) ([]models.CustomerSegmentRevenue, error) {
	query := `
		SELECT 
			COALESCE(c.customer_type, 'unknown') as segment,
			COALESCE(SUM(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as revenue,
			COUNT(DISTINCT c.customer_id) as customer_count,
			COUNT(s.sale_id) as transaction_count,
			COALESCE(AVG(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as avg_transaction_value
		FROM customers c
		LEFT JOIN sales s ON c.customer_id = s.customer_id
			AND s.sale_date BETWEEN $1 AND $2
		GROUP BY c.customer_type
		ORDER BY revenue DESC;
	`

	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer segment revenue: %w", err)
	}
	defer rows.Close()

	var segments []models.CustomerSegmentRevenue
	var totalRevenue float64
	for rows.Next() {
		var csr models.CustomerSegmentRevenue
		if err := rows.Scan(&csr.Segment, &csr.Revenue, &csr.CustomerCount,
			&csr.TransactionCount, &csr.AvgTransactionValue); err != nil {
			return nil, fmt.Errorf("failed to scan customer segment revenue: %w", err)
		}
		totalRevenue += csr.Revenue
		segments = append(segments, csr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over customer segment revenue: %w", err)
	}

	// Calculate metrics
	for i := range segments {
		if totalRevenue > 0 {
			segments[i].RevenuePercentage = (segments[i].Revenue / totalRevenue) * 100
		}
		if segments[i].CustomerCount > 0 {
			segments[i].RevenuePerCustomer = segments[i].Revenue / float64(segments[i].CustomerCount)
		}
	}

	return segments, nil
}

func (dr *DashboardRepository) RegionRevenueSplit(ctx context.Context, startDate, endDate string) ([]models.RegionRevenueSplit, error) {
	query := `
		SELECT 
			r.region_id,
			r.name as region_name,
			COALESCE(SUM(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as revenue,
			COUNT(DISTINCT c.customer_id) as customer_count,
			COUNT(s.sale_id) as transaction_count,
			COALESCE(AVG(COALESCE(s.total_amount, s.unit_price * s.quantity)), 0) as avg_transaction_value
		FROM regions r
		LEFT JOIN customers c ON r.region_id = c.region_id
		LEFT JOIN sales s ON c.customer_id = s.customer_id
			AND s.sale_date BETWEEN $1 AND $2
		GROUP BY r.region_id, r.name
		ORDER BY revenue DESC;
	`

	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve region revenue split: %w", err)
	}
	defer rows.Close()

	var splits []models.RegionRevenueSplit
	var totalRevenue float64
	for rows.Next() {
		var rrs models.RegionRevenueSplit
		if err := rows.Scan(&rrs.RegionID, &rrs.RegionName, &rrs.Revenue,
			&rrs.CustomerCount, &rrs.TransactionCount, &rrs.AvgTransactionValue); err != nil {
			return nil, fmt.Errorf("failed to scan region revenue split: %w", err)
		}
		totalRevenue += rrs.Revenue
		splits = append(splits, rrs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over region revenue split: %w", err)
	}

	// Calculate metrics
	for i := range splits {
		if totalRevenue > 0 {
			splits[i].RevenuePercentage = (splits[i].Revenue / totalRevenue) * 100
		}
		if splits[i].CustomerCount > 0 {
			splits[i].RevenuePerCustomer = splits[i].Revenue / float64(splits[i].CustomerCount)
		}
	}

	return splits, nil
}
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardRepository handles data access for dashboard metrics
type DashboardRepository struct {
	db *pgxpool.Pool
}

// NewDashboardRepository creates a new DashboardRepository
func NewDashboardRepository(db *pgxpool.Pool) (*DashboardRepository, error) {
	repo := &DashboardRepository{db: db}
	// Create necessary tables if they don't exist
	if err := repo.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables required for the dashboard: %w", err)
	}
	return repo, nil
}

// ===== DASHBOARD METHODS =====

// TotalSales retrieves the total sales amount and count for a date range
func (dr *DashboardRepository) TotalSales(ctx context.Context, startDate, endDate string) (float64, int64, error) {
	query := `
		SELECT COALESCE(SUM(total_amount), 0) as total_revenue, COUNT(*) as total_sales
		FROM sales
		WHERE sale_date BETWEEN $1 AND $2;
	`
	var totalRevenue float64
	var totalSales int64
	err := dr.db.QueryRow(ctx, query, startDate, endDate).Scan(&totalRevenue, &totalSales)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to retrieve total sales: %w", err)
	}
	return totalRevenue, totalSales, nil
}

// TargetSales retrieves the target sales amount for a date range and optional region
func (dr *DashboardRepository) TargetSales(ctx context.Context, startDate, endDate string, regionID *uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(target_amount), 0)
		FROM sales_targets st
		JOIN periods p ON st.period_id = p.period_id
		WHERE p.period_start <= $2 AND p.period_end >= $1
	`
	args := []interface{}{startDate, endDate}

	if regionID != nil {
		query += " AND st.region_id = $3"
		args = append(args, *regionID)
	}

	var targetAmount float64
	err := dr.db.QueryRow(ctx, query, args...).Scan(&targetAmount)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve target sales: %w", err)
	}
	return targetAmount, nil
}

// RevenueHistory retrieves revenue data grouped by date for a date range
func (dr *DashboardRepository) RevenueHistory(ctx context.Context, startDate, endDate string) ([]models.RevenueHistory, error) {
	query := `
		SELECT sale_date::DATE as date, COALESCE(SUM(total_amount), 0) as revenue
		FROM sales
		WHERE sale_date BETWEEN $1 AND $2
		GROUP BY sale_date::DATE
		ORDER BY sale_date::DATE;
	`
	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve revenue history: %w", err)
	}
	defer rows.Close()

	var history []models.RevenueHistory
	for rows.Next() {
		var rh models.RevenueHistory
		var date time.Time
		if err := rows.Scan(&date, &rh.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan revenue history: %w", err)
		}
		rh.Date = date
		history = append(history, rh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over revenue history: %w", err)
	}

	return history, nil
}

// SalesConversionRates retrieves conversion rates based on leads for a date range
func (dr *DashboardRepository) SalesConversionRates(ctx context.Context, startDate, endDate string) ([]models.ConversionRate, error) {
	query := `
		SELECT created_at::DATE as date, 
			COUNT(*) as opportunities, 
			COUNT(*) FILTER (WHERE status = 'converted') as conversions
		FROM leads
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY created_at::DATE
		ORDER BY created_at::DATE;
	`
	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve conversion rates: %w", err)
	}
	defer rows.Close()

	var rates []models.ConversionRate
	for rows.Next() {
		var cr models.ConversionRate
		var date time.Time
		var opportunities, conversions int64
		if err := rows.Scan(&date, &opportunities, &conversions); err != nil {
			return nil, fmt.Errorf("failed to scan conversion rate: %w", err)
		}
		cr.Date = date
		cr.Opportunities = opportunities
		cr.Conversions = conversions
		if opportunities > 0 {
			cr.ConversionRate = float64(conversions) / float64(opportunities) * 100
		}
		rates = append(rates, cr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over conversion rates: %w", err)
	}

	return rates, nil
}

// ForecastData retrieves sales forecast data for a date range
func (dr *DashboardRepository) ForecastData(ctx context.Context, startDate, endDate string) ([]models.ForecastData, error) {
	query := `
		SELECT f.period_id, COALESCE(SUM(f.forecasted_amount), 0) as forecast_amount
		FROM forecasts f
		JOIN periods p ON f.period_id = p.period_id
		WHERE p.period_start <= $2 AND p.period_end >= $1
		GROUP BY f.period_id
		ORDER BY f.period_id;
	`
	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve forecast data: %w", err)
	}
	defer rows.Close()

	var forecasts []models.ForecastData
	for rows.Next() {
		var fd models.ForecastData
		var periodID uuid.UUID
		if err := rows.Scan(&periodID, &fd.ForecastAmount); err != nil {
			return nil, fmt.Errorf("failed to scan forecast data: %w", err)
		}
		fd.PeriodID = periodID
		forecasts = append(forecasts, fd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over forecast data: %w", err)
	}

	return forecasts, nil
}

// ProductSummary retrieves sales summary by product
func (dr *DashboardRepository) ProductSummary(ctx context.Context, startDate, endDate string) ([]models.ProductSummary, error) {
	query := `
		SELECT p.product_id, p.name, COALESCE(SUM(s.quantity), 0) as total_quantity, `
		if err := rows.Scan(&rrs.RegionID, &rrs.RegionName, &rrs.Revenue,
			&rrs.CustomerCount, &rrs.TransactionCount, &rrs.AvgTransactionValue); err != nil {
			return nil, fmt.Errorf("failed to scan region revenue split: %w", err)
		}
		totalRevenue += rrs.Revenue
		splits = append(splits, rrs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over region revenue split: %w", err)
	}

	// Calculate metrics
	for i := range splits {
		if totalRevenue > 0 {
			splits[i].RevenuePercentage = (splits[i].Revenue / totalRevenue) * 100
		}
		if splits[i].CustomerCount > 0 {
			splits[i].RevenuePerCustomer = splits[i].Revenue / float64(splits[i].CustomerCount)
		}
	}

	return splits, nil
}

func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
}

func GetRecentNotifications(userID uint, limit int) ([]models.Notification, error) {
	var notifs []models.Notification
	if err := db.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

func MarkNotificationAsRead(notificationID int) error {
	return db.DB.Model(&models.Notification{}).Where("id = ?", notificationID).Update("is_read", true).Error
	Update("is_read", true).Error
}

func SearchSaleMetrics(userID, uint, query string) ([]models.SaleMetric, error) {
	var metrics []models.SaleMetric
	searchQuery := fmt.Sprintf("%%%s%%", query)
	if err := db.DB.Where("user_id = ? AND (product_name ILIKE ? OR category ILIKE ?)", userID, searchQuery, searchQuery).
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}
