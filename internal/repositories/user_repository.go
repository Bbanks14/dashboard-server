package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bbanks14/dashboard-server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyInUse = errors.New("email already in use")
	ErrInvalidUserData   = errors.New("invalid user data")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return ErrInvalidUserData
	}

	const query = `
		INSERT INTO users (
			name, email, password, city, state, country, 
			occupation, phone_number, transactions, role
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at, updated_at
	`

	// Handle nil transactions array
	var transactions interface{}
	if user.Transactions != nil {
		transactions = pq.Array(user.Transactions)
	} else {
		transactions = pq.Array([]string{})
	}

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password,
		user.City,
		user.State,
		user.Country,
		user.Occupation,
		user.PhoneNumber,
		transactions,
		user.Role,
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrEmailAlreadyInUse
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, ErrInvalidUserData
	}

	const query = `
		SELECT id, name, email, city, state, country, occupation, 
		       phone_number, transactions, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var transactions pq.StringArray

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.City,
		&user.State,
		&user.Country,
		&user.Occupation,
		&user.PhoneNumber,
		&transactions,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.Transactions = []string(transactions)
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if email == "" {
		return nil, ErrInvalidUserData
	}

	const query = `
		SELECT id, name, email, password, city, state, country, occupation, 
		       phone_number, transactions, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	var transactions pq.StringArray

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.City,
		&user.State,
		&user.Country,
		&user.Occupation,
		&user.PhoneNumber,
		&transactions,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Transactions = []string(transactions)
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	if user == nil || user.ID == "" {
		return ErrInvalidUserData
	}

	const query = `
		UPDATE users
		SET 
			name = $1,
			email = $2,
			city = $3,
			state = $4,
			country = $5,
			occupation = $6,
			phone_number = $7,
			transactions = $8,
			role = $9,
			updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`

	// Handle transactions array
	var transactions interface{}
	if user.Transactions != nil {
		transactions = pq.Array(user.Transactions)
	} else {
		transactions = pq.Array([]string{})
	}

	args := []interface{}{
		user.Name,
		user.Email,
		user.City,
		user.State,
		user.Country,
		user.Occupation,
		user.PhoneNumber,
		transactions,
		user.Role,
		user.ID,
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(&user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrEmailAlreadyInUse
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserData
	}

	const query = `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, error) {
	if page < 1 || pageSize < 1 {
		return nil, ErrInvalidUserData
	}

	const query = `
		SELECT id, name, email, city, state, country, occupation, 
		       phone_number, transactions, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, pageSize)
	for rows.Next() {
		var user domain.User
		var transactions pq.StringArray

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.City,
			&user.State,
			&user.Country,
			&user.Occupation,
			&user.PhoneNumber,
			&transactions,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Transactions = []string(transactions)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, newPassword string) error {
	if id == "" || newPassword == "" {
		return ErrInvalidUserData
	}

	const query = `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, newPassword, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) AddUserTransaction(ctx context.Context, userID string, transactionID string) error {
	if userID == "" || transactionID == "" {
		return ErrInvalidUserData
	}

	const query = `
		UPDATE users
		SET transactions = array_append(transactions, $1),
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, transactionID, userID)
	if err != nil {
		return fmt.Errorf("failed to add transaction to user: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUserCount returns the total number of users (useful for pagination)
func (r *UserRepository) GetUserCount(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM users`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}

// GetUsersByRole returns users filtered by role
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string, page, pageSize int) ([]domain.User, error) {
	if role == "" || page < 1 || pageSize < 1 {
		return nil, ErrInvalidUserData
	}

	const query = `
		SELECT id, name, email, city, state, country, occupation, 
		       phone_number, transactions, role, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, query, role, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, pageSize)
	for rows.Next() {
		var user domain.User
		var transactions pq.StringArray

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.City,
			&user.State,
			&user.Country,
			&user.Occupation,
			&user.PhoneNumber,
			&transactions,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Transactions = []string(transactions)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
