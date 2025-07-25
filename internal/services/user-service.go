package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/Bbanks14/dashboard-server/internal/domain"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrInvalidInput       = errors.New("invalid input provided")
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// validateEmail performs basic email validation
func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	email = strings.TrimSpace(strings.ToLower(email))

	// Basic email validation - contains @ and has parts before/after
	if !strings.Contains(email, "@") ||
		strings.HasPrefix(email, "@") ||
		strings.HasSuffix(email, "@") {
		return ErrInvalidEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return ErrInvalidEmail
	}

	return nil
}

// validatePassword checks password strength
func (s *UserService) validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}

	return nil
}

// sanitizeInput trims whitespace and validates non-empty strings
func (s *UserService) sanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

func (s *UserService) RegisterUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return ErrInvalidInput
	}

	// Validate and sanitize email
	user.Email = strings.ToLower(s.sanitizeInput(user.Email))
	if err := s.validateEmail(user.Email); err != nil {
		return err
	}

	// Validate password
	if err := s.validatePassword(user.Password); err != nil {
		return err
	}

	// Sanitize other fields
	user.Name = s.sanitizeInput(user.Name)
	if user.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	// Check if email already exists
	existing, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Set default role if not provided
	if user.Role == "" {
		user.Role = domain.RoleUser
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return s.userRepo.CreateUser(ctx, user)
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	// Sanitize and validate inputs
	email = strings.ToLower(s.sanitizeInput(email))
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

func (s *UserService) GetUserProfile(ctx context.Context, id string) (*domain.User, error) {
	if s.sanitizeInput(id) == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Clear sensitive data
	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, id string, update *domain.User) (*domain.User, error) {
	if s.sanitizeInput(id) == "" || update == nil {
		return nil, ErrInvalidInput
	}

	// Fetch existing user
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate and update allowed fields
	if update.Name != "" {
		user.Name = s.sanitizeInput(update.Name)
		if user.Name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
		}
	}

	// Sanitize other optional fields
	user.City = s.sanitizeInput(update.City)
	user.State = s.sanitizeInput(update.State)
	user.Country = s.sanitizeInput(update.Country)
	user.Occupation = s.sanitizeInput(update.Occupation)
	user.PhoneNumber = s.sanitizeInput(update.PhoneNumber)
	user.UpdatedAt = time.Now()

	// If email is being updated, validate and check for uniqueness
	if update.Email != "" {
		newEmail := strings.ToLower(s.sanitizeInput(update.Email))
		if err := s.validateEmail(newEmail); err != nil {
			return nil, err
		}

		if newEmail != user.Email {
			existing, err := s.userRepo.GetUserByEmail(ctx, newEmail)
			if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
				return nil, fmt.Errorf("failed to check email: %w", err)
			}
			if existing != nil {
				return nil, ErrEmailAlreadyExists
			}
			user.Email = newEmail
		}
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	if s.sanitizeInput(id) == "" || currentPassword == "" || newPassword == "" {
		return ErrInvalidInput
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	users, err := s.userRepo.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Clear passwords from all users
	for i := range users {
		users[i].Password = ""
	}

	// Get total count for pagination metadata
	total, err := s.userRepo.GetUserCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return users, total, nil
}

// DeleteUser soft deletes a user (you might want to implement this)
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if s.sanitizeInput(id) == "" {
		return ErrInvalidInput
	}

	return s.userRepo.DeleteUser(ctx, id)
}

// GetUserByEmail retrieves a user by email (useful for admin operations)
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.ToLower(s.sanitizeInput(email))
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Clear sensitive data
	user.Password = ""
	return user, nil
}
