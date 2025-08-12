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

func (dr *DashboardRepository) createTables() error {
	// Create the update_updated_at_column function
	updateFunction := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`

	if _, err := dr.db.Exec(context.Background(), updateFunction); err != nil {
		return fmt.Errorf("failed to create update function: %w", err)
	}

	// Unified table definitions that work for both dashboard and revenue growth
	usersTable := `
		CREATE TABLE IF NOT EXISTS users (
			user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(100) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100),
			last_name VARCHAR(100),
			role VARCHAR(50) DEFAULT 'user',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	regionsTable := `
		CREATE TABLE IF NOT EXISTS regions (
			region_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			manager_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	productsTable := `
		CREATE TABLE IF NOT EXISTS products (
			product_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(200) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2),
			category VARCHAR(100),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Unified customers table with customer_type column
	customersTable := `
		CREATE TABLE IF NOT EXISTS customers (
			customer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(200) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(20),
			address TEXT,
			region_id UUID REFERENCES regions(region_id) ON DELETE SET NULL,
			customer_type VARCHAR(50) DEFAULT 'regular',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Unified periods table with consistent column names
	periodsTable := `
		CREATE TABLE IF NOT EXISTS periods (
			period_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			period_type VARCHAR(50) NOT NULL,
			period_start DATE NOT NULL,
			period_end DATE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Unified sales table with both unit_price and total_amount
	salesTable := `
		CREATE TABLE IF NOT EXISTS sales (
			sale_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID REFERENCES products(product_id) ON DELETE SET NULL,
			customer_id UUID REFERENCES customers(customer_id) ON DELETE SET NULL,
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			unit_price DECIMAL(10, 2) NOT NULL CHECK (unit_price >= 0),
			total_amount DECIMAL(10, 2) NOT NULL CHECK (total_amount >= 0),
			sale_date TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	salesTargetsTable := `
		CREATE TABLE IF NOT EXISTS sales_targets (
			target_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			region_id UUID REFERENCES regions(region_id) ON DELETE SET NULL,
			period_id UUID REFERENCES periods(period_id) ON DELETE CASCADE,
			target_amount DECIMAL(10, 2) NOT NULL CHECK (target_amount >= 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	forecastsTable := `
		CREATE TABLE IF NOT EXISTS forecasts (
			forecast_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			period_id UUID REFERENCES periods(period_id) ON DELETE CASCADE,
			product_id UUID REFERENCES products(product_id) ON DELETE CASCADE,
			region_id UUID REFERENCES regions(region_id) ON DELETE SET NULL,
			forecasted_amount DECIMAL(10, 2) NOT NULL CHECK (forecasted_amount >= 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	leadsTable := `
		CREATE TABLE IF NOT EXISTS leads (
			lead_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(20),
			source VARCHAR(50),
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Create tables in dependency order
	tables := []string{
		usersTable,
		regionsTable,
		productsTable,
		customersTable,
		periodsTable,
		salesTable,
		salesTargetsTable,
		forecastsTable,
		leadsTable,
	}

	// Create tables
	for _, table := range tables {
		if _, err := dr.db.Exec(context.Background(), table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create triggers for updated_at columns
	triggers := []string{
		`DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
		 CREATE TRIGGER trigger_users_updated_at BEFORE UPDATE ON users
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_regions_updated_at ON regions;
		 CREATE TRIGGER trigger_regions_updated_at BEFORE UPDATE ON regions
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_customers_updated_at ON customers;
		 CREATE TRIGGER trigger_customers_updated_at BEFORE UPDATE ON customers
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_products_updated_at ON products;
		 CREATE TRIGGER trigger_products_updated_at BEFORE UPDATE ON products
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_periods_updated_at ON periods;
		 CREATE TRIGGER trigger_periods_updated_at BEFORE UPDATE ON periods
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_sales_updated_at ON sales;
		 CREATE TRIGGER trigger_sales_updated_at BEFORE UPDATE ON sales
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_sales_targets_updated_at ON sales_targets;
		 CREATE TRIGGER trigger_sales_targets_updated_at BEFORE UPDATE ON sales_targets
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_forecasts_updated_at ON forecasts;
		 CREATE TRIGGER trigger_forecasts_updated_at BEFORE UPDATE ON forecasts
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS trigger_leads_updated_at ON leads;
		 CREATE TRIGGER trigger_leads_updated_at BEFORE UPDATE ON leads
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,
	}

	// Create triggers (non-critical, so just log warnings)
	for _, trigger := range triggers {
		if _, err := dr.db.Exec(context.Background(), trigger); err != nil {
			fmt.Printf("Warning: Could not create trigger: %v\n", err)
		}
	}

	return nil
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
		SELECT p.product_id, p.name, COALESCE(SUM(s.quantity), 0) as total_quantity, 
			COALESCE(SUM(s.total_amount), 0) as total_revenue
		FROM products p
		LEFT JOIN sales s ON p.product_id = s.product_id
			AND s.sale_date BETWEEN $1 AND $2
		GROUP BY p.product_id, p.name
		ORDER BY total_revenue DESC;
	`
	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve product summary: %w", err)
	}
	defer rows.Close()

	var summaries []models.ProductSummary
	for rows.Next() {
		var ps models.ProductSummary
		if err := rows.Scan(&ps.ProductID, &ps.ProductName, &ps.UnitsSold, &ps.TotalSales); err != nil {
			return nil, fmt.Errorf("failed to scan product summary: %w", err)
		}
		summaries = append(summaries, ps)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over product summaries: %w", err)
	}

	return summaries, nil
}

// RegionSummary retrieves sales summary by region
func (dr *DashboardRepository) RegionSummary(ctx context.Context, startDate, endDate string) ([]models.RegionSummary, error) {
	query := `
		SELECT r.region_id, r.name, 
			COALESCE(SUM(s.quantity), 0) as total_quantity, 
			COALESCE(SUM(s.total_amount), 0) as total_revenue
		FROM regions r
		LEFT JOIN customers c ON r.region_id = c.region_id
		LEFT JOIN sales s ON c.customer_id = s.customer_id
			AND s.sale_date BETWEEN $1 AND $2
		GROUP BY r.region_id, r.name
		ORDER BY total_revenue DESC;
	`
	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve region summary: %w", err)
	}
	defer rows.Close()

	var summaries []models.RegionSummary
	for rows.Next() {
		var rs models.RegionSummary
		if err := rows.Scan(&rs.RegionID, &rs.RegionName, &rs.TotalQuantity, &rs.TotalSales); err != nil {
			return nil, fmt.Errorf("failed to scan region summary: %w", err)
		}
		summaries = append(summaries, rs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over region summaries: %w", err)
	}

	return summaries, nil
}

// ===== REVENUE GROWTH METHODS =====

func (dr *DashboardRepository) RevenueByDate(ctx context.Context, startDate, endDate string, granularity string) ([]models.RevenueByDate, error) {
	var dateFormat string
	switch granularity {
	case "daily":
		dateFormat = "sale_date::DATE"
	case "weekly":
		dateFormat = "DATE_TRUNC('week', sale_date)::DATE"
	case "monthly":
		dateFormat = "DATE_TRUNC('month', sale_date)::DATE"
	case "quarterly":
		dateFormat = "DATE_TRUNC('quarter', sale_date)::DATE"
	case "yearly":
		dateFormat = "DATE_TRUNC('year', sale_date)::DATE"
	default:
		dateFormat = "sale_date::DATE"
	}

	// Use total_amount if available, otherwise calculate from unit_price * quantity
	query := fmt.Sprintf(`
		SELECT %s as date,
			COALESCE(SUM(COALESCE(total_amount, unit_price * quantity)), 0) as revenue,
			COUNT(*) as transaction_count,
			COALESCE(AVG(COALESCE(total_amount, unit_price * quantity)), 0) as avg_transaction_value
		FROM sales
		WHERE sale_date BETWEEN $1 AND $2
		GROUP BY %s
		ORDER BY %s;
	`, dateFormat, dateFormat, dateFormat)

	rows, err := dr.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve revenue by date: %w", err)
	}
	defer rows.Close()

	var revenues []models.RevenueByDate
	for rows.Next() {
		var rbd models.RevenueByDate
		var date time.Time
		if err := rows.Scan(&date, &rbd.Revenue, &rbd.TransactionCount, &rbd.AvgTransactionValue); err != nil {
			return nil, fmt.Errorf("failed to scan revenue by date: %w", err)
		}
		rbd.Date = date
		revenues = append(revenues, rbd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over revenue by date: %w", err)
	}

	return revenues, nil
}

func (dr *DashboardRepository) PeriodSalesComparison(ctx context.Context, currentPeriodStart, currentPeriodEnd, previousPeriodStart, previousPeriodEnd string) (models.PeriodComparison, error) {
	query := `
		WITH current_period AS (
			SELECT
				COALESCE(SUM(COALESCE(total_amount, unit_price * quantity)), 0) as revenue,
				COUNT(*) as transaction_count,
				COALESCE(AVG(COALESCE(total_amount, unit_price * quantity)), 0) as avg_transaction_value
			FROM sales
			WHERE sale_date BETWEEN $1 AND $2
		),
		previous_period AS (
			SELECT
				COALESCE(SUM(COALESCE(total_amount, unit_price * quantity)), 0) as revenue,
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
