package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents auth user (internal use only)
type User struct {
	ID           string
	Username     string
	PasswordHash string
	IsActive     bool
	Role         string // 'user' or 'super_admin'
}

// AuthRepository handles raw SQL for auth
type AuthRepository struct {
	db *pgxpool.Pool
}

// NewAuthRepository creates a new repo
func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

/*
========================
 CREATE USER
========================
*/
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

/*
========================
 GET USER BY USERNAME
========================
*/
func (r *AuthRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*User, error) {

	const query = `
		SELECT id, username, password_hash, is_active, role
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
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // user not found
		}
		return nil, err
	}

	return &user, nil
}

/*
========================
 UPDATE USER ROLE
========================
*/
func (r *AuthRepository) UpdateUserRole(
	ctx context.Context,
	userID string,
	role string,
) error {

	const query = `
		UPDATE users
		SET role = $1, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, role, userID)
	return err
}
