package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// MonthlyData represents monthly sales data
type MonthlyData struct {
	Month      int     `json:"month"`
	TotalSales float64 `json:"totalSales"`
	TotalUnits int     `json:"totalUnits"`
}

// DailyData represents daily sales data
type DailyData struct {
	Date       time.Time `json:"date"`
	TotalSales float64   `json:"totalSales"`
	TotalUnits int       `json:"totalUnits"`
}

// ProductStat represents the statistics for a product
type ProductStat struct {
	ID                   uuid.UUID     `json:"_id" db:"id"`
	ProductID            uuid.UUID     `json:"productId" db:"product_id"`
	YearlySalesTotal     float64       `json:"yearlySalesTotal" db:"yearly_sales_total"`
	YearlyTotalSoldUnits int           `json:"yearlyTotalSoldUnits" db:"yearly_total_sold_units"`
	Year                 int           `json:"year" db:"year"`
	MonthlyData          []MonthlyData `json:"monthlyData" db:"monthly_data"`
	DailyData            []DailyData   `json:"dailyData" db:"daily_data"`
	CreatedAt            time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time     `json:"updatedAt" db:"updated_at"`
}

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID   `json:"_id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Price       float64     `json:"price" db:"price"`
	Rating      float64     `json:"rating" db:"rating"`
	Category    string      `json:"category" db:"category"`
	Supply      int         `json:"supply" db:"supply"`
	Stat        ProductStat `json:"stat" db:"-"`
	CreatedAt   time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time   `json:"updatedAt" db:"updated_at"`
}

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository instance
func NewProductRepository(db *sql.DB) *ProductRepository {
	repo := &ProductRepository{db: db}

	// Create tables if they don't exist
	if err := repo.createTables(); err != nil {
		panic(fmt.Sprintf("Failed to create tables: %v\n", err))
	}

	return repo
}

// createTables creates the necessary tables if they don't exist
func (pr *ProductRepository) createTables() error {
	// Create products table
	productsTable := `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL DEFAULT 0,
			rating DECIMAL(3,2) NOT NULL DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
			category VARCHAR(100) NOT NULL,
			supply INTEGER NOT NULL DEFAULT 0 CHECK (supply >= 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Create product_stats table
	statsTable := `
		CREATE TABLE IF NOT EXISTS product_stats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			yearly_sales_total DECIMAL(12,2) NOT NULL DEFAULT 0,
			yearly_total_sold_units INTEGER NOT NULL DEFAULT 0,
			year INTEGER NOT NULL DEFAULT EXTRACT(YEAR FROM CURRENT_DATE),
			monthly_data JSONB DEFAULT '[]'::jsonb,
			daily_data JSONB DEFAULT '[]'::jsonb,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(product_id, year)
		);
	`

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);",
		"CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);",
		"CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);",
		"CREATE INDEX IF NOT EXISTS idx_products_rating ON products(rating);",
		"CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_product_stats_product_id ON product_stats(product_id);",
		"CREATE INDEX IF NOT EXISTS idx_product_stats_year ON product_stats(year);",
	}

	// Create trigger for updating updated_at
	trigger := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_products_updated_at ON products;
		CREATE TRIGGER update_products_updated_at
			BEFORE UPDATE ON products
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();

		DROP TRIGGER IF EXISTS update_product_stats_updated_at ON product_stats;
		CREATE TRIGGER update_product_stats_updated_at
			BEFORE UPDATE ON product_stats
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`

	// Execute table creation
	if _, err := pr.db.Exec(productsTable); err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	if _, err := pr.db.Exec(statsTable); err != nil {
		return fmt.Errorf("failed to create product_stats table: %w", err)
	}

	// Create indexes
	for _, index := range indexes {
		if _, err := pr.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create triggers
	if _, err := pr.db.Exec(trigger); err != nil {
		return fmt.Errorf("failed to create triggers: %w", err)
	}

	return nil
}

// GetAllProducts retrieves all products with their statistics
func (pr *ProductRepository) GetAllProducts(ctx context.Context) ([]Product, error) {
	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.rating, p.category, p.supply, 
			p.created_at, p.updated_at,
			COALESCE(ps.id, gen_random_uuid()) as stat_id,
			p.id as stat_product_id,
			COALESCE(ps.yearly_sales_total, 0) as yearly_sales_total,
			COALESCE(ps.yearly_total_sold_units, 0) as yearly_total_sold_units,
			COALESCE(ps.year, EXTRACT(YEAR FROM CURRENT_DATE)::INTEGER) as year,
			COALESCE(ps.monthly_data, '[]'::jsonb) as monthly_data,
			COALESCE(ps.daily_data, '[]'::jsonb) as daily_data,
			COALESCE(ps.created_at, CURRENT_TIMESTAMP) as stat_created_at,
			COALESCE(ps.updated_at, CURRENT_TIMESTAMP) as stat_updated_at
		FROM products p
		LEFT JOIN product_stats ps ON p.id = ps.product_id AND ps.year = EXTRACT(YEAR FROM CURRENT_DATE)
		ORDER BY p.created_at DESC
	`

	rows, err := pr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var monthlyDataJSON, dailyDataJSON []byte

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Rating, &product.Category, &product.Supply,
			&product.CreatedAt, &product.UpdatedAt,
			&product.Stat.ID, &product.Stat.ProductID,
			&product.Stat.YearlySalesTotal, &product.Stat.YearlyTotalSoldUnits,
			&product.Stat.Year, &monthlyDataJSON, &dailyDataJSON,
			&product.Stat.CreatedAt, &product.Stat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse JSON data
		if err := json.Unmarshal(monthlyDataJSON, &product.Stat.MonthlyData); err != nil {
			product.Stat.MonthlyData = []MonthlyData{}
		}
		if err := json.Unmarshal(dailyDataJSON, &product.Stat.DailyData); err != nil {
			product.Stat.DailyData = []DailyData{}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return products, nil
}

// GetProductByID retrieves a single product by ID with its statistics
func (pr *ProductRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.rating, p.category, p.supply, 
			p.created_at, p.updated_at,
			COALESCE(ps.id, gen_random_uuid()) as stat_id,
			p.id as stat_product_id,
			COALESCE(ps.yearly_sales_total, 0) as yearly_sales_total,
			COALESCE(ps.yearly_total_sold_units, 0) as yearly_total_sold_units,
			COALESCE(ps.year, EXTRACT(YEAR FROM CURRENT_DATE)::INTEGER) as year,
			COALESCE(ps.monthly_data, '[]'::jsonb) as monthly_data,
			COALESCE(ps.daily_data, '[]'::jsonb) as daily_data,
			COALESCE(ps.created_at, CURRENT_TIMESTAMP) as stat_created_at,
			COALESCE(ps.updated_at, CURRENT_TIMESTAMP) as stat_updated_at
		FROM products p
		LEFT JOIN product_stats ps ON p.id = ps.product_id AND ps.year = EXTRACT(YEAR FROM CURRENT_DATE)
		WHERE p.id = $1
	`

	var product Product
	var monthlyDataJSON, dailyDataJSON []byte

	err := pr.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Rating, &product.Category, &product.Supply,
		&product.CreatedAt, &product.UpdatedAt,
		&product.Stat.ID, &product.Stat.ProductID,
		&product.Stat.YearlySalesTotal, &product.Stat.YearlyTotalSoldUnits,
		&product.Stat.Year, &monthlyDataJSON, &dailyDataJSON,
		&product.Stat.CreatedAt, &product.Stat.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to query product: %w", err)
	}

	// Parse JSON data
	if err := json.Unmarshal(monthlyDataJSON, &product.Stat.MonthlyData); err != nil {
		product.Stat.MonthlyData = []MonthlyData{}
	}
	if err := json.Unmarshal(dailyDataJSON, &product.Stat.DailyData); err != nil {
		product.Stat.DailyData = []DailyData{}
	}

	return &product, nil
}

// CreateProduct creates a new product
func (pr *ProductRepository) CreateProduct(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (name, description, price, rating, category, supply)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := pr.db.QueryRowContext(ctx, query,
		product.Name, product.Description, product.Price,
		product.Rating, product.Category, product.Supply,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// UpdateProduct updates an existing product
func (pr *ProductRepository) UpdateProduct(ctx context.Context, id uuid.UUID, product *Product) error {
	query := `
		UPDATE products 
		SET name = $2, description = $3, price = $4, rating = $5, category = $6, supply = $7
		WHERE id = $1
		RETURNING updated_at
	`

	err := pr.db.QueryRowContext(ctx, query, id,
		product.Name, product.Description, product.Price,
		product.Rating, product.Category, product.Supply,
	).Scan(&product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	product.ID = id
	return nil
}

// DeleteProduct deletes a product by ID
func (pr *ProductRepository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := pr.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetProductsByCategory retrieves products by category
func (pr *ProductRepository) GetProductsByCategory(ctx context.Context, category string) ([]Product, error) {
	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.rating, p.category, p.supply, 
			p.created_at, p.updated_at,
			COALESCE(ps.id, gen_random_uuid()) as stat_id,
			p.id as stat_product_id,
			COALESCE(ps.yearly_sales_total, 0) as yearly_sales_total,
			COALESCE(ps.yearly_total_sold_units, 0) as yearly_total_sold_units,
			COALESCE(ps.year, EXTRACT(YEAR FROM CURRENT_DATE)::INTEGER) as year,
			COALESCE(ps.monthly_data, '[]'::jsonb) as monthly_data,
			COALESCE(ps.daily_data, '[]'::jsonb) as daily_data,
			COALESCE(ps.created_at, CURRENT_TIMESTAMP) as stat_created_at,
			COALESCE(ps.updated_at, CURRENT_TIMESTAMP) as stat_updated_at
		FROM products p
		LEFT JOIN product_stats ps ON p.id = ps.product_id AND ps.year = EXTRACT(YEAR FROM CURRENT_DATE)
		WHERE p.category = $1
		ORDER BY p.created_at DESC
	`

	rows, err := pr.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by category: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var monthlyDataJSON, dailyDataJSON []byte

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Rating, &product.Category, &product.Supply,
			&product.CreatedAt, &product.UpdatedAt,
			&product.Stat.ID, &product.Stat.ProductID,
			&product.Stat.YearlySalesTotal, &product.Stat.YearlyTotalSoldUnits,
			&product.Stat.Year, &monthlyDataJSON, &dailyDataJSON,
			&product.Stat.CreatedAt, &product.Stat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse JSON data
		if err := json.Unmarshal(monthlyDataJSON, &product.Stat.MonthlyData); err != nil {
			product.Stat.MonthlyData = []MonthlyData{}
		}
		if err := json.Unmarshal(dailyDataJSON, &product.Stat.DailyData); err != nil {
			product.Stat.DailyData = []DailyData{}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return products, nil
}

// SearchProducts searches for products by name or description
func (pr *ProductRepository) SearchProducts(ctx context.Context, query string) ([]Product, error) {
	sqlQuery := `
		SELECT 
			p.id, p.name, p.description, p.price, p.rating, p.category, p.supply, 
			p.created_at, p.updated_at,
			COALESCE(ps.id, gen_random_uuid()) as stat_id,
			p.id as stat_product_id,
			COALESCE(ps.yearly_sales_total, 0) as yearly_sales_total,
			COALESCE(ps.yearly_total_sold_units, 0) as yearly_total_sold_units,
			COALESCE(ps.year, EXTRACT(YEAR FROM CURRENT_DATE)::INTEGER) as year,
			COALESCE(ps.monthly_data, '[]'::jsonb) as monthly_data,
			COALESCE(ps.daily_data, '[]'::jsonb) as daily_data,
			COALESCE(ps.created_at, CURRENT_TIMESTAMP) as stat_created_at,
			COALESCE(ps.updated_at, CURRENT_TIMESTAMP) as stat_updated_at
		FROM products p
		LEFT JOIN product_stats ps ON p.id = ps.product_id AND ps.year = EXTRACT(YEAR FROM CURRENT_DATE)
		WHERE p.name ILIKE $1 OR p.description ILIKE $1
		ORDER BY p.created_at DESC
	`

	searchPattern := "%" + query + "%"
	rows, err := pr.db.QueryContext(ctx, sqlQuery, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var monthlyDataJSON, dailyDataJSON []byte

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Rating, &product.Category, &product.Supply,
			&product.CreatedAt, &product.UpdatedAt,
			&product.Stat.ID, &product.Stat.ProductID,
			&product.Stat.YearlySalesTotal, &product.Stat.YearlyTotalSoldUnits,
			&product.Stat.Year, &monthlyDataJSON, &dailyDataJSON,
			&product.Stat.CreatedAt, &product.Stat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse JSON data
		if err := json.Unmarshal(monthlyDataJSON, &product.Stat.MonthlyData); err != nil {
			product.Stat.MonthlyData = []MonthlyData{}
		}
		if err := json.Unmarshal(dailyDataJSON, &product.Stat.DailyData); err != nil {
			product.Stat.DailyData = []DailyData{}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return products, nil
}

// GetProductsWithPagination retrieves products with pagination
func (pr *ProductRepository) GetProductsWithPagination(ctx context.Context, page, limit int) ([]Product, int64, error) {
	// Get total count
	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM products`
	err := pr.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.rating, p.category, p.supply, 
			p.created_at, p.updated_at,
			COALESCE(ps.id, gen_random_uuid()) as stat_id,
			p.id as stat_product_id,
			COALESCE(ps.yearly_sales_total, 0) as yearly_sales_total,
			COALESCE(ps.yearly_total_sold_units, 0) as yearly_total_sold_units,
			COALESCE(ps.year, EXTRACT(YEAR FROM CURRENT_DATE)::INTEGER) as year,
			COALESCE(ps.monthly_data, '[]'::jsonb) as monthly_data,
			COALESCE(ps.daily_data, '[]'::jsonb) as daily_data,
			COALESCE(ps.created_at, CURRENT_TIMESTAMP) as stat_created_at,
			COALESCE(ps.updated_at, CURRENT_TIMESTAMP) as stat_updated_at
		FROM products p
		LEFT JOIN product_stats ps ON p.id = ps.product_id AND ps.year = EXTRACT(YEAR FROM CURRENT_DATE)
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := pr.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products with pagination: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var monthlyDataJSON, dailyDataJSON []byte

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Rating, &product.Category, &product.Supply,
			&product.CreatedAt, &product.UpdatedAt,
			&product.Stat.ID, &product.Stat.ProductID,
			&product.Stat.YearlySalesTotal, &product.Stat.YearlyTotalSoldUnits,
			&product.Stat.Year, &monthlyDataJSON, &dailyDataJSON,
			&product.Stat.CreatedAt, &product.Stat.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse JSON data
		if err := json.Unmarshal(monthlyDataJSON, &product.Stat.MonthlyData); err != nil {
			product.Stat.MonthlyData = []MonthlyData{}
		}
		if err := json.Unmarshal(dailyDataJSON, &product.Stat.DailyData); err != nil {
			product.Stat.DailyData = []DailyData{}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return products, totalCount, nil
}

// UpdateProductSupply updates the supply of a product
func (pr *ProductRepository) UpdateProductSupply(ctx context.Context, id uuid.UUID, newSupply int) error {
	query := `UPDATE products SET supply = $2 WHERE id = $1`

	result, err := pr.db.ExecContext(ctx, query, id, newSupply)
	if err != nil {
		return fmt.Errorf("failed to update product supply: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// CreateProductStat creates or updates product statistics
func (pr *ProductRepository) CreateProductStat(ctx context.Context, stat *ProductStat) error {
	monthlyDataJSON, err := json.Marshal(stat.MonthlyData)
	if err != nil {
		return fmt.Errorf("failed to marshal monthly data: %w", err)
	}

	dailyDataJSON, err := json.Marshal(stat.DailyData)
	if err != nil {
		return fmt.Errorf("failed to marshal daily data: %w", err)
	}

	query := `
		INSERT INTO product_stats (product_id, yearly_sales_total, yearly_total_sold_units, year, monthly_data, daily_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = pr.db.QueryRowContext(ctx, query,
		stat.ProductID, stat.YearlySalesTotal, stat.YearlyTotalSoldUnits,
		stat.Year, monthlyDataJSON, dailyDataJSON,
	).Scan(&stat.ID, &stat.CreatedAt, &stat.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product stat: %w", err)
	}

	return nil
}

// UpdateProductStat updates product statistics
func (pr *ProductRepository) UpdateProductStat(ctx context.Context, productID uuid.UUID, stat *ProductStat) error {
	monthlyDataJSON, err := json.Marshal(stat.MonthlyData)
	if err != nil {
		return fmt.Errorf("failed to marshal monthly data: %w", err)
	}

	dailyDataJSON, err := json.Marshal(stat.DailyData)
	if err != nil {
		return fmt.Errorf("failed to marshal daily data: %w", err)
	}

	query := `
		INSERT INTO product_stats (product_id, yearly_sales_total, yearly_total_sold_units, year, monthly_data, daily_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (product_id, year)
		DO UPDATE SET 
			yearly_sales_total = EXCLUDED.yearly_sales_total,
			yearly_total_sold_units = EXCLUDED.yearly_total_sold_units,
			monthly_data = EXCLUDED.monthly_data,
			daily_data = EXCLUDED.daily_data
		RETURNING id, created_at, updated_at
	`

	err = pr.db.QueryRowContext(ctx, query,
		productID, stat.YearlySalesTotal, stat.YearlyTotalSoldUnits,
		stat.Year, monthlyDataJSON, dailyDataJSON,
	).Scan(&stat.ID, &stat.CreatedAt, &stat.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update product stat: %w", err)
	}

	stat.ProductID = productID
	return nil
}
