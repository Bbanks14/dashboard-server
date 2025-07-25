package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Bbanks14/dashboard-server/internal/util/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database handles database connections and operations
type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase creates a new database connection pool
func NewDatabase(config config.DatabaseConfig) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// AffiliateModel handles affiliate-related database operations
type AffiliateModel struct {
	pool *pgxpool.Pool
}

// CreateTableIfNotExists ensures the affiliates table exists
func (m *AffiliateModel) CreateTableIfNotExists() error {
	query := `
	CREATE TABLE IF NOT EXISTS affiliates (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		commission DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);`

	_, err := m.pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// GetClients retrieves clients with pagination, sorting, and search
func (db *Database) GetClients(page, pageSize int, sort, search string) (interface{}, int, error) {
	ctx := context.Background()
	offset := (page - 1) * pageSize

	baseQuery := `
		SELECT id, company_name, contact_name, contact_email, phone, address, 
		       city, state, country, created_at, updated_at
		FROM clients
	`
	countQuery := "SELECT COUNT(*) FROM clients"

	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = ` WHERE (
			company_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			contact_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			contact_email ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}

	orderClause := fmt.Sprintf(" ORDER BY %s", db.sanitizeSortField(sort))
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Execute count query
	var totalCount int
	countArgs := args[:len(args)-2] // Remove LIMIT and OFFSET args for count
	err := db.Pool.QueryRow(ctx, countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get clients count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.Pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []map[string]interface{}
	for rows.Next() {
		var id int
		var companyName, contactName, contactEmail, phone, address, city, state, country string
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &companyName, &contactName, &contactEmail, &phone,
			&address, &city, &state, &country, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan client row: %w", err)
		}

		client := map[string]interface{}{
			"id":            id,
			"company_name":  companyName,
			"contact_name":  contactName,
			"contact_email": contactEmail,
			"phone":         phone,
			"address":       address,
			"city":          city,
			"state":         state,
			"country":       country,
			"created_at":    createdAt.Time,
			"updated_at":    updatedAt.Time,
		}
		clients = append(clients, client)
	}

	return clients, totalCount, nil
}

// GetProducts retrieves products with pagination, sorting, and search
// Returns products in format compatible with frontend expectations
func (db *Database) GetProducts(page, pageSize int, sort, search string) (interface{}, int, error) {
	ctx := context.Background()
	offset := (page - 1) * pageSize

	// Join with product_stats table to match original JavaScript functionality
	baseQuery := `
		SELECT p.id, p.name, p.description, p.price, p.category, p.stock_quantity, 
		       p.sku, p.brand, p.is_active, p.created_at, p.updated_at,
		       ps.yearly_sales_total, ps.yearly_units_sold, ps.monthly_data, ps.daily_data
		FROM products p
		LEFT JOIN product_stats ps ON p.id = ps.product_id
	`
	countQuery := "SELECT COUNT(*) FROM products p"

	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = ` WHERE (
			p.name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			p.description ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			p.category ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			p.brand ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}

	orderClause := fmt.Sprintf(" ORDER BY p.%s", db.sanitizeSortField(sort))
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Get total count
	var totalCount int
	countArgs := args[:len(args)-2]
	err := db.Pool.QueryRow(ctx, countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.Pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int
		var name, description, category, sku, brand string
		var price float64
		var stockQuantity int
		var isActive bool
		var createdAt, updatedAt sql.NullTime
		var yearlySalesTotal, yearlyUnitsSold sql.NullFloat64
		var monthlyData, dailyData sql.NullString

		err := rows.Scan(&id, &name, &description, &price, &category, &stockQuantity,
			&sku, &brand, &isActive, &createdAt, &updatedAt,
			&yearlySalesTotal, &yearlyUnitsSold, &monthlyData, &dailyData)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}

		// Build stat object to match original JavaScript structure
		stat := map[string]interface{}{
			"productId":        id,
			"yearlySalesTotal": yearlySalesTotal.Float64,
			"yearlyUnitsSold":  yearlyUnitsSold.Float64,
			"monthlyData":      monthlyData.String,
			"dailyData":        dailyData.String,
		}

		product := map[string]interface{}{
			"_id":            fmt.Sprintf("%d", id), // Convert to string to match MongoDB ObjectId format
			"name":           name,
			"description":    description,
			"price":          price,
			"category":       category,
			"stock_quantity": stockQuantity,
			"sku":            sku,
			"brand":          brand,
			"is_active":      isActive,
			"created_at":     createdAt.Time,
			"updated_at":     updatedAt.Time,
			"stat":           []map[string]interface{}{stat}, // Array format to match original
		}
		products = append(products, product)
	}

	return products, totalCount, nil
}

// GetCustomers retrieves users with role="user" (customers only) with pagination, sorting, and search
func (db *Database) GetCustomers(page, pageSize int, sort, search string) (interface{}, int, error) {
	ctx := context.Background()
	offset := (page - 1) * pageSize

	baseQuery := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, 
		       address, city, state, country, occupation, role, created_at, updated_at
		FROM users
		WHERE role = 'user'
	`
	countQuery := "SELECT COUNT(*) FROM users WHERE role = 'user'"

	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = ` AND (
			first_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			last_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			email ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			occupation ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}

	orderClause := fmt.Sprintf(" ORDER BY %s", db.sanitizeSortField(sort))
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Get total count
	var totalCount int
	countArgs := args[:len(args)-2]
	err := db.Pool.QueryRow(ctx, countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get customers count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.Pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []map[string]interface{}
	for rows.Next() {
		var id int
		var firstName, lastName, email, phone, address, city, state, country, occupation, role string
		var dateOfBirth sql.NullTime
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &firstName, &lastName, &email, &phone, &dateOfBirth,
			&address, &city, &state, &country, &occupation, &role, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer row: %w", err)
		}

		customer := map[string]interface{}{
			"_id":           fmt.Sprintf("%d", id), // String format to match MongoDB
			"first_name":    firstName,
			"last_name":     lastName,
			"email":         email,
			"phone":         phone,
			"date_of_birth": dateOfBirth.Time,
			"address":       address,
			"city":          city,
			"state":         state,
			"country":       country,
			"occupation":    occupation,
			"role":          role,
			"created_at":    createdAt.Time,
			"updated_at":    updatedAt.Time,
		}
		customers = append(customers, customer)
	}

	return customers, totalCount, nil
}

// GetTransactions retrieves transactions with pagination, sorting, and search
func (db *Database) GetTransactions(page, pageSize int, sort, search string) (interface{}, int, error) {
	ctx := context.Background()
	offset := (page - 1) * pageSize

	baseQuery := `
		SELECT t.id, t.user_id, t.product_id, t.quantity, t.unit_price, t.total_amount,
		       t.transaction_date, t.status, t.payment_method, t.created_at,
		       u.first_name, u.last_name, u.email,
		       p.name as product_name, p.category
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN products p ON t.product_id = p.id
	`
	countQuery := `
		SELECT COUNT(*) 
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN products p ON t.product_id = p.id
	`

	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = ` WHERE (
			CAST(t.total_amount AS TEXT) ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			CAST(t.user_id AS TEXT) ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			u.first_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			u.last_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			u.email ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			p.name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			t.status ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Fix sort field to handle table prefixes properly
	sortField := db.sanitizeSortFieldForTransactions(sort)
	orderClause := fmt.Sprintf(" ORDER BY %s", sortField)
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Get total count
	var totalCount int
	countArgs := args[:len(args)-2]
	err := db.Pool.QueryRow(ctx, countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.Pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id, userID, productID, quantity int
		var unitPrice, totalAmount float64
		var transactionDate sql.NullTime
		var status, paymentMethod string
		var createdAt sql.NullTime
		var firstName, lastName, email, productName, category sql.NullString

		err := rows.Scan(&id, &userID, &productID, &quantity, &unitPrice, &totalAmount,
			&transactionDate, &status, &paymentMethod, &createdAt,
			&firstName, &lastName, &email, &productName, &category)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		transaction := map[string]interface{}{
			"_id":              fmt.Sprintf("%d", id),            // String format to match MongoDB
			"userId":           fmt.Sprintf("%d", userID),        // Match original field name
			"cost":             fmt.Sprintf("%.2f", totalAmount), // String format to match original search
			"products":         []string{productName.String},     // Array format to match original
			"quantity":         quantity,
			"unit_price":       unitPrice,
			"total_amount":     totalAmount,
			"transaction_date": transactionDate.Time,
			"status":           status,
			"payment_method":   paymentMethod,
			"created_at":       createdAt.Time,
			"user": map[string]interface{}{
				"first_name": firstName.String,
				"last_name":  lastName.String,
				"email":      email.String,
			},
			"product": map[string]interface{}{
				"name":     productName.String,
				"category": category.String,
			},
		}
		transactions = append(transactions, transaction)
	}

	return transactions, totalCount, nil
}

// GetGeography retrieves user count by location for geography charts
// Returns data in format expected by frontend: [{id: "USA", value: 25}]
func (db *Database) GetGeography() (interface{}, error) {
	ctx := context.Background()

	query := `
		SELECT country, COUNT(*) as user_count
		FROM users 
		WHERE country IS NOT NULL AND country != '' AND role = 'user'
		GROUP BY country
		ORDER BY user_count DESC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by location: %w", err)
	}
	defer rows.Close()

	var locations []map[string]interface{}
	for rows.Next() {
		var country string
		var userCount int

		err := rows.Scan(&country, &userCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan location row: %w", err)
		}

		// Convert country name to ISO3 code to match original JavaScript functionality
		countryISO3 := db.convertCountryToISO3(country)

		location := map[string]interface{}{
			"id":    countryISO3, // Match original format
			"value": userCount,   // Match original format
		}
		locations = append(locations, location)
	}

	return locations, nil
}

// convertCountryToISO3 converts full country names to ISO3 codes
// This is a simplified version - you may want to use a proper library
func (db *Database) convertCountryToISO3(countryName string) string {
	countryMap := map[string]string{
		"United States":        "USA",
		"Canada":               "CAN",
		"United Kingdom":       "GBR",
		"Germany":              "DEU",
		"France":               "FRA",
		"Australia":            "AUS",
		"Japan":                "JPN",
		"China":                "CHN",
		"India":                "IND",
		"Brazil":               "BRA",
		"Mexico":               "MEX",
		"Italy":                "ITA",
		"Spain":                "ESP",
		"Netherlands":          "NLD",
		"Sweden":               "SWE",
		"Norway":               "NOR",
		"Denmark":              "DNK",
		"Finland":              "FIN",
		"Poland":               "POL",
		"Russia":               "RUS",
		"South Korea":          "KOR",
		"Singapore":            "SGP",
		"New Zealand":          "NZL",
		"South Africa":         "ZAF",
		"Argentina":            "ARG",
		"Chile":                "CHL",
		"Colombia":             "COL",
		"Peru":                 "PER",
		"Venezuela":            "VEN",
		"Turkey":               "TUR",
		"Israel":               "ISR",
		"Saudi Arabia":         "SAU",
		"United Arab Emirates": "ARE",
		"Egypt":                "EGY",
		"Nigeria":              "NGA",
		"Kenya":                "KEN",
		"Ghana":                "GHA",
		"Morocco":              "MAR",
		"Thailand":             "THA",
		"Vietnam":              "VNM",
		"Philippines":          "PHL",
		"Indonesia":            "IDN",
		"Malaysia":             "MYS",
		"Pakistan":             "PAK",
		"Bangladesh":           "BGD",
		"Sri Lanka":            "LKA",
	}

	if iso3, exists := countryMap[countryName]; exists {
		return iso3
	}

	// If not found, return first 3 letters of country name as fallback
	if len(countryName) >= 3 {
		return strings.ToUpper(countryName[:3])
	}
	return strings.ToUpper(countryName)
}

// sanitizeSortField ensures only safe column names are used for sorting
func (db *Database) sanitizeSortField(sort string) string {
	allowedFields := map[string]string{
		"id":           "id",
		"name":         "name",
		"email":        "email",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
		"price":        "price",
		"quantity":     "quantity",
		"first_name":   "first_name",
		"last_name":    "last_name",
		"company_name": "company_name",
		"contact_name": "contact_name",
	}

	if field, exists := allowedFields[sort]; exists {
		return field
	}
	return "id" // Default sort field
}

// sanitizeSortFieldForTransactions ensures only safe column names are used for sorting transactions
func (db *Database) sanitizeSortFieldForTransactions(sort string) string {
	allowedFields := map[string]string{
		"id":         "t.id",
		"date":       "t.transaction_date",
		"created_at": "t.created_at",
		"updated_at": "t.updated_at",
		"total":      "t.total_amount",
		"amount":     "t.total_amount",
		"quantity":   "t.quantity",
		"status":     "t.status",
		"user_id":    "t.user_id",
		"product_id": "t.product_id",
	}

	if field, exists := allowedFields[sort]; exists {
		return field
	}
	return "t.id" // Default sort field with table prefix
}
