package forms

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FormsRepository handles DB operations for forms
type FormsRepository struct {
	db *pgxpool.Pool
}

// NewFormsRepository creates repo
func NewFormsRepository(db *pgxpool.Pool) *FormsRepository {
	return &FormsRepository{db: db}
}

/*
========================
 CREATE
========================
*/
func (r *FormsRepository) Create(
	ctx context.Context,
	userID string,
	title string,
	description string,
) (*Form, error) {

	const query = `
		INSERT INTO forms (user_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, status, created_at, updated_at
	`

	var f Form
	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		title,
		description,
	).Scan(
		&f.ID,
		&f.Title,
		&f.Description,
		&f.Status,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &f, nil
}

/*
========================
 GET BY ID
========================
*/
func (r *FormsRepository) GetByID(
	ctx context.Context,
	userID string,
	formID string,
) (*Form, error) {

	const query = `
		SELECT id, title, description, status, auto_slug, custom_slug, accepting_responses, published_at, is_template, created_at, updated_at
		FROM forms
		WHERE
			id = $1
			AND user_id = $2
			AND deleted_at IS NULL
	`

	var f Form
	err := r.db.QueryRow(
		ctx,
		query,
		formID,
		userID,
	).Scan(
		&f.ID,
		&f.Title,
		&f.Description,
		&f.Status,
		&f.AutoSlug,
		&f.CustomSlug,
		&f.AcceptingResponses,
		&f.PublishedAt,
		&f.IsTemplate,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &f, nil
}

/*
========================
 LIST
========================
*/
func (r *FormsRepository) List(
	ctx context.Context,
	userID string,
	search string,
	limit int,
	offset int,
) ([]Form, int, error) {

	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var (
		rows  pgx.Rows
		err   error
		forms []Form
		total int
	)

	if search == "" {
		// -------- NO SEARCH --------
		listQuery := `
			SELECT id, title, description, status, auto_slug, custom_slug, accepting_responses, published_at, is_template, created_at, updated_at
			FROM forms
			WHERE user_id = $1
			  AND deleted_at IS NULL
			ORDER BY updated_at DESC
			LIMIT $2 OFFSET $3
		`

		rows, err = r.db.Query(ctx, listQuery, userID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

		countQuery := `
			SELECT COUNT(*)
			FROM forms
			WHERE user_id = $1
			  AND deleted_at IS NULL
		`

		err = r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}

	} else {
		// -------- WITH SEARCH --------
		listQuery := `
			SELECT id, title, description, status, auto_slug, custom_slug, accepting_responses, published_at, is_template, created_at, updated_at
			FROM forms
			WHERE user_id = $1
			  AND deleted_at IS NULL
			  AND title ILIKE '%' || $2 || '%'
			ORDER BY updated_at DESC
			LIMIT $3 OFFSET $4
		`

		rows, err = r.db.Query(ctx, listQuery, userID, search, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

		countQuery := `
			SELECT COUNT(*)
			FROM forms
			WHERE user_id = $1
			  AND deleted_at IS NULL
			  AND title ILIKE '%' || $2 || '%'
		`

		err = r.db.QueryRow(ctx, countQuery, userID, search).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	for rows.Next() {
		var f Form
		if err := rows.Scan(
			&f.ID,
			&f.Title,
			&f.Description,
			&f.Status,
			&f.AutoSlug,
			&f.CustomSlug,
			&f.AcceptingResponses,
			&f.PublishedAt,
			&f.IsTemplate,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		forms = append(forms, f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return forms, total, nil
}


/*
========================
 UPDATE
========================
*/
func (r *FormsRepository) Update(
	ctx context.Context,
	userID string,
	formID string,
	title string,
	description string,
	status string,
) error {

	const query = `
		UPDATE forms
		SET
			title = $1,
			description = $2,
			status = $3,
			updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		WHERE
			id = $4
			AND user_id = $5
			AND deleted_at IS NULL
	`

	cmd, err := r.db.Exec(
		ctx,
		query,
		title,
		description,
		status,
		formID,
		userID,
	)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

/*
========================
 SOFT DELETE
========================
*/
func (r *FormsRepository) SoftDelete(
	ctx context.Context,
	userID string,
	formID string,
) error {

	const query = `
		UPDATE forms
		SET deleted_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		WHERE
			id = $1
			AND user_id = $2
			AND deleted_at IS NULL
	`

	cmd, err := r.db.Exec(ctx, query, formID, userID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

/*
========================
 GET FORM SLUGS
========================
*/
func (r *FormsRepository) GetFormSlugs(
	ctx context.Context,
	formID string,
) (*string, *string, error) {
	var autoSlug, customSlug *string

	const query = `
		SELECT auto_slug, custom_slug
		FROM forms
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, formID).Scan(&autoSlug, &customSlug)
	if err == pgx.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	return autoSlug, customSlug, nil
}

/*
========================
 TEMPLATE OPERATIONS
========================
*/

// ToggleTemplate toggles is_template flag (super admin only)
func (r *FormsRepository) ToggleTemplate(
	ctx context.Context,
	formID string,
	isTemplate bool,
) error {
	const query = `
		UPDATE forms
		SET
			is_template = $1,
			updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		WHERE
			id = $2
			AND deleted_at IS NULL
			AND status = 'published'
	`

	cmd, err := r.db.Exec(ctx, query, isTemplate, formID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ListTemplates lists all published templates (public)
func (r *FormsRepository) ListTemplates(
	ctx context.Context,
) ([]Form, error) {
	const query = `
		SELECT id, title, description, status, is_template, created_at, updated_at
		FROM forms
		WHERE is_template = true
		  AND deleted_at IS NULL
		  AND status = 'published'
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []Form
	for rows.Next() {
		var f Form
		if err := rows.Scan(
			&f.ID,
			&f.Title,
			&f.Description,
			&f.Status,
			&f.IsTemplate,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			continue
		}
		forms = append(forms, f)
	}

	return forms, nil
}

// GetTemplateWithFlow returns template metadata (for cloning)
func (r *FormsRepository) GetTemplateWithFlow(
	ctx context.Context,
	templateID string,
) (*Form, error) {
	const query = `
		SELECT id, title, description, status, is_template, created_at, updated_at
		FROM forms
		WHERE id = $1 AND is_template = true AND status = 'published' AND deleted_at IS NULL
	`

	var f Form
	err := r.db.QueryRow(ctx, query, templateID).Scan(
		&f.ID,
		&f.Title,
		&f.Description,
		&f.Status,
		&f.IsTemplate,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if err != nil {
		return nil, ErrNotFound
	}

	return &f, nil
}
