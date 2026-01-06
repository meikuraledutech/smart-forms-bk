package auth

import (
	"context"
	"errors"
	"os"
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

// LoginResponse contains login result
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// UserResponse contains user info for client
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login validates credentials and issues tokens
func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (*LoginResponse, error) {

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Do NOT reveal whether username exists
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if !VerifyPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Bootstrap super admin from ENV
	superAdminUsername := os.Getenv("SUPER_ADMIN_USERNAME")
	if superAdminUsername != "" && username == superAdminUsername && user.Role != "super_admin" {
		// Promote this user to super admin
		err = s.repo.UpdateUserRole(ctx, user.ID, "super_admin")
		if err == nil {
			user.Role = "super_admin"
		}
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	}, nil
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

	// Use role from existing token (avoids DB query)
	// If role is empty (old token), default to 'user'
	role := claims.Role
	if role == "" {
		role = "user"
	}

	return GenerateAccessToken(claims.UserID, role)
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
