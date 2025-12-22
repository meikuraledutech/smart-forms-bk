package auth

import (
	"context"
	"errors"
)

/*
========================
 DOMAIN ERRORS
========================
*/

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

/*
========================
 SERVICE
========================
*/

// AuthService coordinates auth logic
type AuthService struct {
	repo *AuthRepository
}

// NewAuthService creates auth service
func NewAuthService(repo *AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

/*
========================
 LOGIN
========================
*/

// Login validates credentials and issues tokens
func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (accessToken string, refreshToken string, err error) {

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	// Do NOT reveal whether username exists
	if user == nil {
		return "", "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", "", ErrUserInactive
	}

	if !VerifyPassword(password, user.PasswordHash) {
		return "", "", ErrInvalidCredentials
	}

	accessToken, err = GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

/*
========================
 REFRESH TOKEN
========================
*/

// RefreshAccessToken issues a new access token
func (s *AuthService) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (string, error) {

	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// TODO (future):
	// - Check Redis blacklist
	// - Token rotation
	// - Device/session validation

	return GenerateAccessToken(claims.UserID)
}

/*
========================
 REGISTER
========================
*/

// Register creates a new user
func (s *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) error {

	// Check if user already exists
	existing, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserAlreadyExists
	}

	// Hash password
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Create user
	return s.repo.CreateUser(ctx, username, hash)
}
