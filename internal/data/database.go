package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bbanks14/dashboard-backend/internal/util/config"
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

	pool, err := pgxpool.log.Println(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{Pool: pool}, nil
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
	// Calculate offset for pagination
	offset := (page - 1) * pageSize

	// Base query
	baseQuery := `
		SELECT id, company_name, contact_name, contact_email, phone, address, 
		       city, state, country, created_at, updated_at
		FROM clients
	`

	countQuery := "SELECT COUNT(*) FROM clients"

	// Build WHERE clause for search
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

	// Build ORDER BY clause
	orderClause := fmt.Sprintf(" ORDER BY %s", db.sanitizeSortField(sort))

	// Build LIMIT and OFFSET clause
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Execute count query
	var totalCount int
	countArgs := args[:len(args)-2] // Remove LIMIT and OFFSET args for count
	err := db.conn.QueryRow(countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get clients count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.conn.Query(fullQuery, args...)
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
func (db *Database) GetProducts(page, pageSize int, sort, search string) (interface{}, int, error) {
	offset := (page - 1) * pageSize

	baseQuery := `
		SELECT id, name, description, price, category, stock_quantity, 
		       sku, brand, is_active, created_at, updated_at
		FROM products
	`

	countQuery := "SELECT COUNT(*) FROM products"

	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = ` WHERE (
			name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			description ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			category ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			brand ILIKE $` + fmt.Sprintf("%d", argIndex) + `
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
	err := db.conn.QueryRow(countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.conn.Query(fullQuery, args...)
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

		err := rows.Scan(&id, &name, &description, &price, &category, &stockQuantity,
			&sku, &brand, &isActive, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}

		product := map[string]interface{}{
			"id":             id,
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
		}
		products = append(products, product)
	}

	return products, totalCount, nil
}

// GetUsers retrieves users with pagination, sorting, and search
func (db *Database) GetUsers(page, pageSize int, sort, search string) (interface{}, int, error) {
	offset := (page - 1) * pageSize

	baseQuery := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, 
		       address, city, state, country, occupation, created_at, updated_at
		FROM users
	`

	countQuery := "SELECT COUNT(*) FROM users"

	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = ` WHERE (
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
	err := db.conn.QueryRow(countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.conn.Query(fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var firstName, lastName, email, phone, address, city, state, country, occupation string
		var dateOfBirth sql.NullTime
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &firstName, &lastName, &email, &phone, &dateOfBirth,
			&address, &city, &state, &country, &occupation, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user row: %w", err)
		}

		user := map[string]interface{}{
			"id":            id,
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
			"created_at":    createdAt.Time,
			"updated_at":    updatedAt.Time,
		}
		users = append(users, user)
	}

	return users, totalCount, nil
}

// GetTransactions retrieves transactions with pagination, sorting, and search
func (db *Database) GetTransactions(page, pageSize int, sort, search string) (interface{}, int, error) {
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
			u.first_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			u.last_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			u.email ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			p.name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			t.status ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}

	orderClause := fmt.Sprintf(" ORDER BY t.%s", db.sanitizeSortField(sort))
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Get total count
	var totalCount int
	countArgs := args[:len(args)-2]
	err := db.conn.QueryRow(countQuery+whereClause, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions count: %w", err)
	}

	// Execute main query
	fullQuery := baseQuery + whereClause + orderClause + limitClause
	rows, err := db.conn.Query(fullQuery, args...)
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
			"id":               id,
			"user_id":          userID,
			"product_id":       productID,
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

// GetUsersByLocation retrieves user count by location for geography charts
func (db *Database) GetUsersByLocation() (interface{}, error) {
	query := `
		SELECT country, COUNT(*) as user_count
		FROM users 
		WHERE country IS NOT NULL AND country != ''
		GROUP BY country
		ORDER BY user_count DESC
	`

	rows, err := db.conn.Query(query)
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

		location := map[string]interface{}{
			"country":    country,
			"user_count": userCount,
		}
		locations = append(locations, location)
	}

	return locations, nil
}

// sanitizeSortField ensures only safe column names are used for sorting
func (db *Database) sanitizeSortField(sort string) string {
	// Define allowed sort fields to prevent SQL injection
	allowedFields := map[string]string{
		"id":         "id",
		"name":       "name",
		"email":      "email",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"price":      "price",
		"quantity":   "quantity",
		"total":      "total_amount",
		"date":       "transaction_date",
		"status":     "status",
	}

	if field, exists := allowedFields[sort]; exists {
		return field
	}

	// Default sort field
	return "id"
}
