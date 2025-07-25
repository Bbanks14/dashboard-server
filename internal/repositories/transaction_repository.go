// File: internal/repositories/transaction_repository.go
package repositories

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/Bbanks14/dashboard-server/internal/domain"
	"github.com/Bbanks14/dashboard-server/internal/data"
)

type transactionRepository struct {
	db *sqlx.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(database *data.Database) TransactionRepository {
	return &transactionRepository{
		db: database.DB,
	}
}

// GetAll retrieves all transactions with pagination
func (r *transactionRepository) GetAll(limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.customer_id, t.products, t.cost, t.created_at,
			u.name as user_name, u.email as user_email,
			c.name as customer_name, c.email as customer_email, c.country
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN customers c ON t.customer_id = c.id
		ORDER BY t.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	var transactions []domain.Transaction
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.Transaction
		var user domain.User
		var customer domain.Customer
		var userName, userEmail, customerName, customerEmail, country sql.NullString

		err := rows.Scan(
			&t.ID, &t.UserID, &t.CustomerID, &t.Products, &t.Cost, &t.CreatedAt,
			&userName, &userEmail, &customerName, &customerEmail, &country,
		)
		if err != nil {
			return nil, err
		}

		// Populate user info if available
		if userName.Valid {
			user.Name = userName.String
			user.Email = userEmail.String
			t.User = &user
		}

		// Populate customer info if available
		if customerName.Valid {
			customer.Name = customerName.String
			customer.Email = customerEmail.String
			customer.Country = country.String
			t.Customer = &customer
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetByID retrieves a transaction by ID
func (r *transactionRepository) GetByID(id uuid.UUID) (*domain.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.customer_id, t.products, t.cost, t.created_at,
			u.name as user_name, u.email as user_email,
			c.name as customer_name, c.email as customer_email, c.country
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN customers c ON t.customer_id = c.id
		WHERE t.id = $1
	`

	var t domain.Transaction
	var user domain.User
	var customer domain.Customer
	var userName, userEmail, customerName, customerEmail, country sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.UserID, &t.CustomerID, &t.Products, &t.Cost, &t.CreatedAt,
		&userName, &userEmail, &customerName, &customerEmail, &country,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Populate user info if available
	if userName.Valid {
		user.Name = userName.String
		user.Email = userEmail.String
		t.User = &user
	}

	// Populate customer info if available
	if customerName.Valid {
		customer.Name = customerName.String
		customer.Email = customerEmail.String
		customer.Country = country.String
		t.Customer = &customer
	}

	return &t, nil
}

// GetByUserID retrieves transactions by user ID
func (r *transactionRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.customer_id, t.products, t.cost, t.created_at,
			c.name as customer_name, c.email as customer_email, c.country
		FROM transactions t
		LEFT JOIN customers c ON t.customer_id = c.id
		WHERE t.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var transactions []domain.Transaction
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.Transaction
		var customer domain.Customer
		var customerName, customerEmail, country sql.NullString

		err := rows.Scan(
			&t.ID, &t.UserID, &t.CustomerID, &t.Products, &t.Cost, &t.CreatedAt,
			&customerName, &customerEmail, &country,
		)
		if err != nil {
			return nil, err
		}

		// Populate customer info if available
		if customerName.Valid {
			customer.Name = customerName.String
			customer.Email = customerEmail.String
			customer.Country = country.String
			t.Customer = &customer
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetByCustomerID retrieves transactions by customer ID
func (r *transactionRepository) GetByCustomerID(customerID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.customer_id, t.products, t.cost, t.created_at,
			u.name as user_name, u.email as user_email
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		WHERE t.customer_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var transactions []domain.Transaction
	rows, err := r.db.Query(query, customerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.Transaction
		var user domain.User
		var userName, userEmail sql.NullString

		err := rows.Scan(
			&t.ID, &t.UserID, &t.CustomerID, &t.Products, &t.Cost, &t.CreatedAt,
			&userName, &userEmail,
		)
		if err != nil {
			return nil, err
		}

		// Populate user info if available
		if userName.Valid {
			user.Name = userName.String
			user.Email = userEmail.String
			t.User = &user
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetByDateRange retrieves transactions within a date range
func (r *transactionRepository) GetByDateRange(startDate, endDate time.Time, limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.customer_id, t.products, t.cost, t.created_at,
			u.name as user_name, u.email as user_email,
			c.name as customer_name, c.email as customer_email, c.country
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN customers c ON t.customer_id = c.id
		WHERE t.created_at >= $1 AND t.created_at <= $2
		ORDER BY t.created_at DESC
		LIMIT $3 OFFSET $4
	`

	var transactions []domain.Transaction
	rows, err := r.db.Query(query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.Transaction
		var user domain.User
		var customer domain.Customer
		var userName, userEmail, customerName, customerEmail, country sql.NullString

		err := rows.Scan(
			&t.ID, &t.UserID, &t.CustomerID, &t.Products, &t.Cost, &t.CreatedAt,
			&userName, &userEmail, &customerName, &customerEmail, &country,
		)
		if err != nil {
			return nil, err
		}

		// Populate user info if available
		if userName.Valid {
			user.Name = userName.String
			user.Email = userEmail.String
			t.User = &user
		}

		// Populate customer info if available
		if customerName.Valid {
			customer.Name = customerName.String
			customer.Email = customerEmail.String
			customer.Country = country.String
			t.Customer = &customer
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetRecent retrieves the most recent transactions
func (r *transactionRepository) GetRecent(limit int) ([]domain.Transaction, error) {
	return r.GetAll(limit, 0)
}

// GetTotalCount retrieves the total count of transactions
func (r *transactionRepository) GetTotalCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	return count, err
}

// Statistics methods

// GetDailyTotals calculates daily sales totals
func (r *transactionRepository) GetDailyTotals(date time.Time) (totalSales float64, totalUnits int, err error) {
	query := `
		SELECT 
			COALESCE(SUM(cost), 0) as total_sales,
			COALESCE(COUNT(*), 0) as total_units
		FROM transactions 
		WHERE DATE(created_at) = DATE($1)
	`
	
	err = r.db.QueryRow(query, date).Scan(&totalSales, &totalUnits)
	return totalSales, totalUnits, err
}

// GetMonthlyTotals calculates monthly sales totals
func (r *transactionRepository) GetMonthlyTotals(year, month int) (totalSales float64, totalUnits int, err error) {
	query := `
		SELECT 
			COALESCE(SUM(cost), 0) as total_sales,
			COALESCE(COUNT(*), 0) as total_units
		FROM transactions 
		WHERE EXTRACT(YEAR FROM created_at) = $1 
		AND EXTRACT(MONTH FROM created_at) = $2
	`
	
	err = r.db.QueryRow(query, year, month).Scan(&totalSales, &totalUnits)
	return totalSales, totalUnits, err
}

// GetYearlyTotals calculates yearly sales totals
func (r *transactionRepository) GetYearlyTotals(year int) (totalSales float64, totalUnits int, err error) {
	query := `
		SELECT 
			CO
