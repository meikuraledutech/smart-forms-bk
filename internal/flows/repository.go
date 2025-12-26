package flows

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FlowRepository struct {
	db *pgxpool.Pool
}

func NewFlowRepository(db *pgxpool.Pool) *FlowRepository {
	return &FlowRepository{db: db}
}

func (r *FlowRepository) DeleteByFormID(ctx context.Context, formID string) error {
	_, err := r.db.Exec(ctx, `UPDATE flow_connections SET deleted_at = NOW() WHERE form_id = $1 AND deleted_at IS NULL`, formID)
	return err
}

func (r *FlowRepository) Create(ctx context.Context, formID, questionID string, parentID *string, orderIndex, depthLevel int, isTerminal bool) (*FlowConnection, error) {
	var fc FlowConnection
	err := r.db.QueryRow(ctx, `
		INSERT INTO flow_connections (form_id, question_id, parent_id, order_index, depth_level, is_terminal)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, form_id, question_id, parent_id, order_index, depth_level, is_terminal, created_at, updated_at
	`, formID, questionID, parentID, orderIndex, depthLevel, isTerminal).Scan(
		&fc.ID, &fc.FormID, &fc.QuestionID, &fc.ParentID, &fc.OrderIndex, &fc.DepthLevel, &fc.IsTerminal, &fc.CreatedAt, &fc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &fc, nil
}

func (r *FlowRepository) GetByFormID(ctx context.Context, formID string) ([]FlowConnection, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, form_id, question_id, parent_id, order_index, depth_level, is_terminal, created_at, updated_at
		FROM flow_connections
		WHERE form_id = $1 AND deleted_at IS NULL
		ORDER BY depth_level, order_index
	`, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []FlowConnection
	for rows.Next() {
		var fc FlowConnection
		err := rows.Scan(&fc.ID, &fc.FormID, &fc.QuestionID, &fc.ParentID, &fc.OrderIndex, &fc.DepthLevel, &fc.IsTerminal, &fc.CreatedAt, &fc.UpdatedAt)
		if err != nil {
			continue
		}
		connections = append(connections, fc)
	}
	return connections, nil
}

func (r *FlowRepository) CreateQuestion(ctx context.Context, userID, qType, text string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO questions (type, question_text, created_by)
		VALUES ($1, $2, $3)
		RETURNING id
	`, qType, text, userID).Scan(&id)
	return id, err
}

func (r *FlowRepository) FindQuestionByText(ctx context.Context, qType, text string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		SELECT id FROM questions WHERE type = $1 AND question_text = $2 AND deleted_at IS NULL LIMIT 1
	`, qType, text).Scan(&id)
	return id, err
}
