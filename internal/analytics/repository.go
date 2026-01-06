package analytics

import (
	"context"
	"encoding/json"
	"smart-forms/internal/analytics/calculators"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// Status methods

func (r *AnalyticsRepository) GetStatus(ctx context.Context, formID string) (*AnalyticsStatus, error) {
	var status AnalyticsStatus

	err := r.db.QueryRow(ctx, `
		SELECT form_id, status, calculated_at, triggered_by, created_at, updated_at
		FROM analytics_status
		WHERE form_id = $1
	`, formID).Scan(&status.FormID, &status.Status, &status.CalculatedAt,
		&status.TriggeredBy, &status.CreatedAt, &status.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (r *AnalyticsRepository) CreateStatus(ctx context.Context, formID, userID, status string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO analytics_status (form_id, status, triggered_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (form_id) DO UPDATE
		SET status = $2, triggered_by = $3, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
	`, formID, status, userID)

	return err
}

func (r *AnalyticsRepository) UpdateStatus(ctx context.Context, formID, status string) error {
	var calculatedAt *string
	if status == "completed" {
		now := "CURRENT_TIMESTAMP AT TIME ZONE 'UTC'"
		calculatedAt = &now
	}

	query := `
		UPDATE analytics_status
		SET status = $1, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
	`

	if calculatedAt != nil {
		query += `, calculated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')`
	}

	query += ` WHERE form_id = $2`

	_, err := r.db.Exec(ctx, query, status, formID)
	return err
}

// Node metrics methods

func (r *AnalyticsRepository) GetNodeMetrics(ctx context.Context, formID string) ([]NodeMetrics, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			an.flow_connection_id,
			q.question_text,
			q.type,
			an.visit_count,
			an.answer_count,
			an.skip_count,
			an.drop_off_count,
			an.total_time_spent,
			an.avg_time_spent,
			an.calculated_at
		FROM analytics_nodes an
		JOIN flow_connections fc ON an.flow_connection_id = fc.id
		JOIN questions q ON fc.question_id = q.id
		WHERE an.form_id = $1
		ORDER BY an.visit_count DESC
	`, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []NodeMetrics
	for rows.Next() {
		var m NodeMetrics
		m.FormID = formID

		err := rows.Scan(&m.FlowConnectionID, &m.QuestionText, &m.QuestionType,
			&m.VisitCount, &m.AnswerCount, &m.SkipCount, &m.DropOffCount,
			&m.TotalTimeSpent, &m.AvgTimeSpent, &m.CalculatedAt)
		if err != nil {
			continue
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// TODO: OPTIMIZE - Sequential Inserts Problem!
// Currently: Executes 100+ individual INSERT queries for 100 metrics
// Solution: Use PostgreSQL batch insert or pgx.CopyFrom
// Example: INSERT INTO analytics_nodes VALUES ($1, $2, ...), ($3, $4, ...), ... (batch 100 rows)
// Expected improvement: 50x faster (100 queries → 1 batch query)
func (r *AnalyticsRepository) SaveNodeMetrics(ctx context.Context, metrics []NodeMetrics) error {
	if len(metrics) == 0 {
		return nil
	}

	// Use batch insert
	for _, m := range metrics {
		_, err := r.db.Exec(ctx, `
			INSERT INTO analytics_nodes (form_id, flow_connection_id, visit_count,
			    answer_count, skip_count, drop_off_count, total_time_spent, avg_time_spent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (form_id, flow_connection_id) DO UPDATE
			SET visit_count = $3, answer_count = $4, skip_count = $5,
			    drop_off_count = $6, total_time_spent = $7, avg_time_spent = $8,
			    calculated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		`, m.FormID, m.FlowConnectionID, m.VisitCount, m.AnswerCount,
			m.SkipCount, m.DropOffCount, m.TotalTimeSpent, m.AvgTimeSpent)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *AnalyticsRepository) DeleteNodeMetrics(ctx context.Context, formID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM analytics_nodes WHERE form_id = $1`, formID)
	return err
}

// Path metrics methods

func (r *AnalyticsRepository) GetPathMetrics(ctx context.Context, formID string) ([]PathMetrics, error) {
	// TODO: Implement
	// SELECT * FROM analytics_paths WHERE form_id = $1
	return nil, nil
}

func (r *AnalyticsRepository) SavePathMetrics(ctx context.Context, metrics []PathMetrics) error {
	// TODO: Implement
	// Batch INSERT INTO analytics_paths
	return nil
}

func (r *AnalyticsRepository) DeletePathMetrics(ctx context.Context, formID string) error {
	// TODO: Implement
	// DELETE FROM analytics_paths WHERE form_id = $1
	return nil
}

// Raw data queries for calculations

// TODO: OPTIMIZE - N+1 Query Problem!
// Currently: For 1000 responses, executes 1001 queries (1 for responses + 1000 for answers)
// Solution: Use single JOIN query or batch fetch answers
// Example: SELECT r.*, a.* FROM form_responses r LEFT JOIN response_answers a ON r.id = a.response_id WHERE r.form_id = $1
// Expected improvement: 1000x faster (1001 queries → 1 query)
func (r *AnalyticsRepository) GetResponseData(ctx context.Context, formID string) ([]calculators.ResponseData, error) {
	// Get all responses for the form
	rows, err := r.db.Query(ctx, `
		SELECT id, flow_path, total_time_spent
		FROM form_responses
		WHERE form_id = $1
		ORDER BY submitted_at DESC
	`, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []calculators.ResponseData
	for rows.Next() {
		var resp calculators.ResponseData
		var flowPathJSON []byte

		err := rows.Scan(&resp.ResponseID, &flowPathJSON, &resp.TotalTimeSpent)
		if err != nil {
			continue
		}

		// Parse flow_path JSON
		if err := json.Unmarshal(flowPathJSON, &resp.FlowPath); err != nil {
			continue
		}

		// Get answers for this response
		answerRows, err := r.db.Query(ctx, `
			SELECT flow_connection_id, answer_text, time_spent
			FROM response_answers
			WHERE response_id = $1
		`, resp.ResponseID)
		if err != nil {
			continue
		}

		var answers []calculators.AnswerData
		for answerRows.Next() {
			var ans calculators.AnswerData
			err := answerRows.Scan(&ans.FlowConnectionID, &ans.AnswerText, &ans.TimeSpent)
			if err != nil {
				continue
			}
			answers = append(answers, ans)
		}
		answerRows.Close()

		resp.Answers = answers
		responses = append(responses, resp)
	}

	return responses, nil
}

// EnrichFlowTransitions adds question text to flow transitions
// TODO: OPTIMIZE - N+1 Query Problem!
// Currently: Executes ~100+ individual queries for 50 transitions
// Solution: Fetch all question texts in ONE query using IN clause or JOIN
// Example: SELECT fc.id, q.question_text FROM flow_connections fc JOIN questions q WHERE fc.id IN (...)
// Expected improvement: 100x faster (100 queries → 1 query)
func (r *AnalyticsRepository) EnrichFlowTransitions(ctx context.Context, transitions []calculators.FlowTransition) ([]FlowTransition, error) {
	if len(transitions) == 0 {
		return []FlowTransition{}, nil
	}

	// Map to cache question texts
	nodeTexts := make(map[string]string)

	enriched := make([]FlowTransition, len(transitions))
	for i, t := range transitions {
		// Get source text
		sourceText := t.SourceID
		if _, exists := nodeTexts[t.SourceID]; !exists {
			var questionText string
			err := r.db.QueryRow(ctx, `
				SELECT q.question_text
				FROM flow_connections fc
				JOIN questions q ON fc.question_id = q.id
				WHERE fc.id = $1
			`, t.SourceID).Scan(&questionText)
			if err == nil {
				nodeTexts[t.SourceID] = questionText
			} else {
				nodeTexts[t.SourceID] = t.SourceID // Fallback to ID
			}
		}
		sourceText = nodeTexts[t.SourceID]

		// Get target text
		targetText := t.TargetID
		if t.IsDropOff {
			targetText = "Drop-off"
		} else {
			if _, exists := nodeTexts[t.TargetID]; !exists {
				var questionText string
				err := r.db.QueryRow(ctx, `
					SELECT q.question_text
					FROM flow_connections fc
					JOIN questions q ON fc.question_id = q.id
					WHERE fc.id = $1
				`, t.TargetID).Scan(&questionText)
				if err == nil {
					nodeTexts[t.TargetID] = questionText
				} else {
					nodeTexts[t.TargetID] = t.TargetID // Fallback to ID
				}
			}
			targetText = nodeTexts[t.TargetID]
		}

		enriched[i] = FlowTransition{
			Source: sourceText,
			Target: targetText,
			Value:  t.Value,
		}
	}

	return enriched, nil
}
