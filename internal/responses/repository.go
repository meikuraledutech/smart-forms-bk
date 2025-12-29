package responses

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ResponsesRepository struct {
	db *pgxpool.Pool
}

func NewResponsesRepository(db *pgxpool.Pool) *ResponsesRepository {
	return &ResponsesRepository{db: db}
}

// GetFormBySlug retrieves form info by slug
func (r *ResponsesRepository) GetFormBySlug(ctx context.Context, slug string) (string, bool, error) {
	var formID string
	var acceptingResponses bool
	err := r.db.QueryRow(ctx, `
		SELECT id, accepting_responses
		FROM forms
		WHERE (auto_slug = $1 OR custom_slug = $1)
		  AND status = 'published'
		  AND deleted_at IS NULL
	`, slug).Scan(&formID, &acceptingResponses)

	if err != nil {
		return "", false, err
	}
	return formID, acceptingResponses, nil
}

// VerifyFlowConnection checks if flow_connection_id exists for the form
func (r *ResponsesRepository) VerifyFlowConnection(ctx context.Context, formID, flowConnectionID string) error {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM flow_connections
			WHERE id = $1 AND form_id = $2 AND deleted_at IS NULL
		)
	`, flowConnectionID, formID).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return ErrInvalidFlowConnection
	}

	return nil
}

// CreateResponse creates a new form response
func (r *ResponsesRepository) CreateResponse(ctx context.Context, formID string, totalTimeSpent int, flowPath []string, metadata map[string]interface{}) (string, error) {
	flowPathJSON, _ := json.Marshal(flowPath)
	metadataJSON, _ := json.Marshal(metadata)

	var responseID string
	err := r.db.QueryRow(ctx, `
		INSERT INTO form_responses (form_id, total_time_spent, flow_path, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, formID, totalTimeSpent, flowPathJSON, metadataJSON).Scan(&responseID)

	return responseID, err
}

// CreateAnswer creates a response answer
func (r *ResponsesRepository) CreateAnswer(ctx context.Context, responseID, flowConnectionID, answerText string, answerValue map[string]interface{}, timeSpent *int) error {
	var answerValueJSON []byte
	if answerValue != nil {
		answerValueJSON, _ = json.Marshal(answerValue)
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO response_answers (response_id, flow_connection_id, answer_text, answer_value, time_spent)
		VALUES ($1, $2, $3, $4, $5)
	`, responseID, flowConnectionID, answerText, answerValueJSON, timeSpent)

	return err
}

// GetResponsesByFormID retrieves all responses for a form
func (r *ResponsesRepository) GetResponsesByFormID(ctx context.Context, formID string, limit, offset int) ([]FormResponse, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, form_id, submitted_at, total_time_spent, flow_path, metadata
		FROM form_responses
		WHERE form_id = $1
		ORDER BY submitted_at DESC
		LIMIT $2 OFFSET $3
	`, formID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var responses []FormResponse
	for rows.Next() {
		var r FormResponse
		var flowPathJSON, metadataJSON []byte

		err := rows.Scan(&r.ID, &r.FormID, &r.SubmittedAt, &r.TotalTimeSpent, &flowPathJSON, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(flowPathJSON, &r.FlowPath)
		json.Unmarshal(metadataJSON, &r.Metadata)

		responses = append(responses, r)
	}

	// Get total count
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM form_responses WHERE form_id = $1`, formID).Scan(&total)

	return responses, total, nil
}

// GetResponseByID retrieves a single response by ID
func (r *ResponsesRepository) GetResponseByID(ctx context.Context, responseID string) (*FormResponse, error) {
	var resp FormResponse
	var flowPathJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, form_id, submitted_at, total_time_spent, flow_path, metadata
		FROM form_responses
		WHERE id = $1
	`, responseID).Scan(&resp.ID, &resp.FormID, &resp.SubmittedAt, &resp.TotalTimeSpent, &flowPathJSON, &metadataJSON)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(flowPathJSON, &resp.FlowPath)
	json.Unmarshal(metadataJSON, &resp.Metadata)

	return &resp, nil
}

// GetAnswersByResponseID retrieves all answers for a response
func (r *ResponsesRepository) GetAnswersByResponseID(ctx context.Context, responseID string) ([]ResponseAnswer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, response_id, flow_connection_id, answer_text, answer_value, time_spent
		FROM response_answers
		WHERE response_id = $1
	`, responseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []ResponseAnswer
	for rows.Next() {
		var a ResponseAnswer
		var answerValueJSON []byte

		err := rows.Scan(&a.ID, &a.ResponseID, &a.FlowConnectionID, &a.AnswerText, &answerValueJSON, &a.TimeSpent)
		if err != nil {
			continue
		}

		if len(answerValueJSON) > 0 {
			json.Unmarshal(answerValueJSON, &a.AnswerValue)
		}

		answers = append(answers, a)
	}

	return answers, nil
}
