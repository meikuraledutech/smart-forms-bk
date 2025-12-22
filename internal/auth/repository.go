package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// User represents auth user (internal use only)
type User struct {
	ID           string
	Username     string
	PasswordHash string
	IsActive     bool
}

// AuthRepository handles raw SQL for auth
type AuthRepository struct {
	db *pgx.Conn
}

// NewAuthRepository creates a new repo
func NewAuthRepository(db *pgx.Conn) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateUser inserts a new user
func (r *AuthRepository) CreateUser(
	ctx context.Context,
	username string,
	passwordHash string,
) error {

	const query = `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, query, username, passwordHash)
	return err
}

// GetUserByUsername fetches user by username
func (r *AuthRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*User, error) {

	const query = `
		SELECT id, username, password_hash, is_active
		FROM users
		WHERE username = $1
	`

	row := r.db.QueryRow(ctx, query, username)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // user not found (important)
		}
		return nil, err
	}

	return &user, nil
}
