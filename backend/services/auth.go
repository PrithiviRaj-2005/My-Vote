package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"pulsevote/models"
	"pulsevote/repository"
	"pulsevote/utils"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Signup validates registration input, hashes password, saves user, and issues a JWT token
func (s *AuthService) Signup(ctx context.Context, req models.SignupRequest) (*models.AuthResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	// Validation: Required fields
	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	// Validation: Email format
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("invalid email address format")
	}

	// Validation: Minimum password length
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	// Validation: Password match if confirm password is provided
	if req.ConfirmPassword != "" && req.Password != req.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}

	// Prevent duplicate email
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("an account with this email already exists")
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email, user.Name)
	if err != nil {
		return nil, errors.New("failed to generate authentication token")
	}

	return &models.AuthResponse{
		Token: token,
		User: models.UserInfo{
			ID:    user.ID.Hex(),
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// Login validates user credentials and returns a JWT token upon success
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare bcrypt password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email, user.Name)
	if err != nil {
		return nil, errors.New("failed to generate authentication token")
	}

	return &models.AuthResponse{
		Token: token,
		User: models.UserInfo{
			ID:    user.ID.Hex(),
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}
