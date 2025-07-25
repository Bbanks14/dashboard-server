package seeds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedInitialData seeds the database with initial data if tables are empty
func SeedInitialData(ctx context.Context, pool *pgxpool.Pool) error {
	// Check if users table is empty
	var userCount int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if userCount > 0 {
		log.Println("Users table already has data, skipping seeding")
		return nil
	}

	// Start transaction
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Seed users
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, google_id, email, name, avatar_url)
		VALUES 
			('550e8400-e29b-41d4-a716-446655440000', 'google_123', 'user1@example.com', 'John Doe', 'https://example.com/avatar1.jpg'),
			('6ba7b810-9dad-11d1-80b4-00c04fd430c8', 'google_456', 'user2@example.com', 'Jane Smith', 'https://example.com/avatar2.jpg')
	`)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Seed products
	_, err = tx.Exec(ctx, `
		INSERT INTO products (id, name, price, description, category, rating, supply)
		VALUES 
			('6ba7b810-9dad-11d1-80b4-00c04fd430c1', 'Product A', 29.99, 'Premium product', 'Electronics', 4.5, 100),
			('6ba7b810-9dad-11d1-80b4-00c04fd430c2', 'Product B', 49.99, 'Luxury item', 'Home', 4.8, 50)
	`)
	if err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	// Seed transactions
	productsJSON, _ := json.Marshal([]string{"6ba7b810-9dad-11d1-80b4-00c04fd430c1"})
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (user_id, cost, products, created_at)
		VALUES 
			('550e8400-e29b-41d4-a716-446655440000', 29.99, $1, $2)
	`, productsJSON, time.Now().Add(-24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to seed transactions: %w", err)
	}

	// Seed overall stats
	currentYear := time.Now().Year()
	_, err = tx.Exec(ctx, `
		INSERT INTO overall_stats (year, total_sales, total_units, avg_sale_price)
		VALUES 
			($1, 150000.00, 2500, 60.00),
			($2, 180000.00, 3000, 60.00)
	`, currentYear-1, currentYear)
	if err != nil {
		return fmt.Errorf("failed to seed overall_stats: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit seed transaction: %w", err)
	}

	log.Println("Initial data seeded successfully")
	return nil
}
