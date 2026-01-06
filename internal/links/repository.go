package links

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LinksRepository struct {
	db *pgxpool.Pool
}

func NewLinksRepository(db *pgxpool.Pool) *LinksRepository {
	return &LinksRepository{db: db}
}

// PublishForm updates form to published status with slugs
func (r *LinksRepository) PublishForm(ctx context.Context, formID, userID, autoSlug string, customSlug *string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE forms
		SET status = 'published',
		    auto_slug = $1,
		    custom_slug = $2,
		    accepting_responses = true,
		    published_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
	`, autoSlug, customSlug, formID, userID)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		return ErrFormNotFound
	}

	return nil
}

// ToggleAcceptingResponses toggles the accepting_responses field
func (r *LinksRepository) ToggleAcceptingResponses(ctx context.Context, formID, userID string, accepting bool) error {
	result, err := r.db.Exec(ctx, `
		UPDATE forms
		SET accepting_responses = $1,
		    updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`, accepting, formID, userID)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		return ErrFormNotFound
	}

	return nil
}

// GetFormBySlug retrieves a published form by auto_slug or custom_slug
func (r *LinksRepository) GetFormBySlug(ctx context.Context, slug string) (string, string, string, bool, error) {
	var formID, title, description string
	var acceptingResponses bool
	err := r.db.QueryRow(ctx, `
		SELECT id, title, description, accepting_responses
		FROM forms
		WHERE (auto_slug = $1 OR custom_slug = $1)
		  AND status = 'published'
		  AND deleted_at IS NULL
	`, slug).Scan(&formID, &title, &description, &acceptingResponses)

	if err != nil {
		return "", "", "", false, err
	}
	return formID, title, description, acceptingResponses, nil
}

// GetFlowForPublicForm retrieves flow structure for a form
func (r *LinksRepository) GetFlowForPublicForm(ctx context.Context, formID string) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			fc.id,
			fc.parent_id,
			fc.order_index,
			q.type,
			q.question_text
		FROM flow_connections fc
		JOIN questions q ON fc.question_id = q.id
		WHERE fc.form_id = $1 AND fc.deleted_at IS NULL
		ORDER BY fc.depth_level, fc.order_index
	`, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, qType, questionText string
		var parentID *string
		var orderIndex int

		err := rows.Scan(&id, &parentID, &orderIndex, &qType, &questionText)
		if err != nil {
			continue
		}

		items = append(items, map[string]interface{}{
			"id":          id,
			"parent_id":   parentID,
			"type":        qType,
			"question":    questionText,
			"order_index": orderIndex,
		})
	}
	return items, nil
}

// CheckSlugExists checks if a slug (auto or custom) is already taken
func (r *LinksRepository) CheckSlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM forms
			WHERE (auto_slug = $1 OR custom_slug = $1)
			  AND deleted_at IS NULL
		)
	`, slug).Scan(&exists)
	return exists, err
}

// GetFormSlugs retrieves the slugs for a form
func (r *LinksRepository) GetFormSlugs(ctx context.Context, formID, userID string) (string, *string, error) {
	var autoSlug string
	var customSlug *string
	err := r.db.QueryRow(ctx, `
		SELECT auto_slug, custom_slug
		FROM forms
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, formID, userID).Scan(&autoSlug, &customSlug)
	return autoSlug, customSlug, err
}
