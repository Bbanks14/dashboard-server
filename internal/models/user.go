
// File: internal/models/user.go
package models

import "time"

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Password  string    `json:"-"` // Password is never sent to client
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository interface for user data operations
type UserRepository interface {
	GetByID(id int64) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id int64) error
	List(limit, offset int) ([]*User, error)
}

// File: internal/services/user_service.go
package services

import (
	"errors"

	"github.com/yourusername/dashboard-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic for users
type UserService struct {
	userRepo models.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo models.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id int64) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// CreateUser creates a new user with password hashing
func (s *UserService) CreateUser(user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	
	return s.userRepo.Create(user)
}

// AuthenticateUser verifies user credentials
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	
	return user, nil
}
