package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	Password     string    `json:"-" db:"password"` // Don't include in JSON responses
	City         string    `json:"city" db:"city"`
	State        string    `json:"state" db:"state"`
	Country      string    `json:"country" db:"country"`
	Occupation   string    `json:"occupation" db:"occupation"`
	PhoneNumber  string    `json:"phone_number" db:"phone_number"`
	Transactions []string  `json:"transactions" db:"transactions"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole constants for different user roles
const (
	UserRoleAdmin  = "admin"
	UserRoleUser   = "user"
	UserRoleClient = "client"
)

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	City        string `json:"city" validate:"max=100"`
	State       string `json:"state" validate:"max=100"`
	Country     string `json:"country" validate:"max=100"`
	Occupation  string `json:"occupation" validate:"max=100"`
	PhoneNumber string `json:"phone_number" validate:"max=20"`
	Role        string `json:"role" validate:"required,oneof=admin user client"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Name        string `json:"name" validate:"min=2,max=100"`
	Email       string `json:"email" validate:"email"`
	City        string `json:"city" validate:"max=100"`
	State       string `json:"state" validate:"max=100"`
	Country     string `json:"country" validate:"max=100"`
	Occupation  string `json:"occupation" validate:"max=100"`
	PhoneNumber string `json:"phone_number" validate:"max=20"`
	Role        string `json:"role" validate:"oneof=admin user client"`
}

// UserResponse represents the response payload for user data (without sensitive info)
type UserResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	Occupation   string    `json:"occupation"`
	PhoneNumber  string    `json:"phone_number"`
	Transactions []string  `json:"transactions"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse (removes sensitive data)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		City:         u.City,
		State:        u.State,
		Country:      u.Country,
		Occupation:   u.Occupation,
		PhoneNumber:  u.PhoneNumber,
		Transactions: u.Transactions,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsClient checks if the user has client role
func (u *User) IsClient() bool {
	return u.Role == UserRoleClient
}

// PaginatedUsersResponse represents paginated user results
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// PasswordUpdateRequest represents the request payload for password updates
type PasswordUpdateRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}
