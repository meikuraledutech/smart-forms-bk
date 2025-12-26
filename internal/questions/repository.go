package questions

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QuestionRepository struct {
	db *pgxpool.Pool
}

func NewQuestionRepository(db *pgxpool.Pool) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) Create(ctx context.Context, userID, qType, text, inputType string, validationRules, metadata map[string]any) (*Question, error) {
	validationJSON, _ := json.Marshal(validationRules)
	metadataJSON, _ := json.Marshal(metadata)

	var q Question
	err := r.db.QueryRow(ctx, `
		INSERT INTO questions (type, question_text, input_type, validation_rules, metadata, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, type, question_text, input_type, validation_rules, metadata, created_by, created_at, updated_at
	`, qType, text, inputType, validationJSON, metadataJSON, userID).Scan(
		&q.ID, &q.Type, &q.QuestionText, &q.InputType, &validationJSON, &metadataJSON, &q.CreatedBy, &q.CreatedAt, &q.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(validationJSON, &q.ValidationRules)
	json.Unmarshal(metadataJSON, &q.Metadata)
	return &q, nil
}

func (r *QuestionRepository) GetByID(ctx context.Context, id string) (*Question, error) {
	var q Question
	var validationJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, type, question_text, input_type, validation_rules, metadata, created_by, created_at, updated_at
		FROM questions WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&q.ID, &q.Type, &q.QuestionText, &q.InputType, &validationJSON, &metadataJSON, &q.CreatedBy, &q.CreatedAt, &q.UpdatedAt,
	)

	if err != nil {
		return nil, ErrNotFound
	}

	json.Unmarshal(validationJSON, &q.ValidationRules)
	json.Unmarshal(metadataJSON, &q.Metadata)
	return &q, nil
}

func (r *QuestionRepository) List(ctx context.Context, qType string, limit, offset int) ([]Question, int, error) {
	var query string
	var args []any

	if qType != "" {
		query = `SELECT id, type, question_text, input_type, validation_rules, metadata, created_by, created_at, updated_at
		         FROM questions WHERE type = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []any{qType, limit, offset}
	} else {
		query = `SELECT id, type, question_text, input_type, validation_rules, metadata, created_by, created_at, updated_at
		         FROM questions WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []any{limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		var validationJSON, metadataJSON []byte
		err := rows.Scan(&q.ID, &q.Type, &q.QuestionText, &q.InputType, &validationJSON, &metadataJSON, &q.CreatedBy, &q.CreatedAt, &q.UpdatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(validationJSON, &q.ValidationRules)
		json.Unmarshal(metadataJSON, &q.Metadata)
		questions = append(questions, q)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM questions WHERE deleted_at IS NULL`
	if qType != "" {
		countQuery = `SELECT COUNT(*) FROM questions WHERE type = $1 AND deleted_at IS NULL`
		r.db.QueryRow(ctx, countQuery, qType).Scan(&total)
	} else {
		r.db.QueryRow(ctx, countQuery).Scan(&total)
	}

	return questions, total, nil
}

func (r *QuestionRepository) Update(ctx context.Context, id, qType, text, inputType string, validationRules, metadata map[string]any) error {
	validationJSON, _ := json.Marshal(validationRules)
	metadataJSON, _ := json.Marshal(metadata)

	_, err := r.db.Exec(ctx, `
		UPDATE questions
		SET type = $1, question_text = $2, input_type = $3, validation_rules = $4, metadata = $5, updated_at = NOW()
		WHERE id = $6
	`, qType, text, inputType, validationJSON, metadataJSON, id)

	if err != nil {
		return ErrNotFound
	}
	return nil
}

func (r *QuestionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `UPDATE questions SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil || result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
