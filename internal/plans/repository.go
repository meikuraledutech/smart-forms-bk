package plans

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlansRepository struct {
	db *pgxpool.Pool
}

func NewPlansRepository(db *pgxpool.Pool) *PlansRepository {
	return &PlansRepository{db: db}
}

// List retrieves all plans (optionally filter by active status)
func (r *PlansRepository) List(ctx context.Context, activeOnly bool) ([]Plan, error) {
	query := `
		SELECT id, name, plan_type, price_inr, razorpay_plan_id, features, is_active, created_at, updated_at
		FROM subscription_plans
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY price_inr ASC"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		var featuresJSON []byte

		err := rows.Scan(&p.ID, &p.Name, &p.PlanType, &p.PriceINR, &p.RazorpayPlanID,
			&featuresJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(featuresJSON, &p.Features); err != nil {
			p.Features = make(map[string]interface{})
		}

		plans = append(plans, p)
	}

	return plans, nil
}

// GetByID retrieves a plan by ID
func (r *PlansRepository) GetByID(ctx context.Context, planID string) (*Plan, error) {
	query := `
		SELECT id, name, plan_type, price_inr, razorpay_plan_id, features, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`

	var p Plan
	var featuresJSON []byte

	err := r.db.QueryRow(ctx, query, planID).Scan(
		&p.ID, &p.Name, &p.PlanType, &p.PriceINR, &p.RazorpayPlanID,
		&featuresJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(featuresJSON, &p.Features); err != nil {
		p.Features = make(map[string]interface{})
	}

	return &p, nil
}

// Create creates a new plan
func (r *PlansRepository) Create(ctx context.Context, req CreatePlanRequest) (*Plan, error) {
	featuresJSON, _ := json.Marshal(req.Features)

	query := `
		INSERT INTO subscription_plans (name, plan_type, price_inr, razorpay_plan_id, features)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, plan_type, price_inr, razorpay_plan_id, features, is_active, created_at, updated_at
	`

	var p Plan
	var returnedFeaturesJSON []byte

	err := r.db.QueryRow(ctx, query, req.Name, req.PlanType, req.PriceINR,
		req.RazorpayPlanID, featuresJSON).Scan(
		&p.ID, &p.Name, &p.PlanType, &p.PriceINR, &p.RazorpayPlanID,
		&returnedFeaturesJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(returnedFeaturesJSON, &p.Features); err != nil {
		p.Features = make(map[string]interface{})
	}

	return &p, nil
}

// Update updates a plan
func (r *PlansRepository) Update(ctx context.Context, planID string, req UpdatePlanRequest) (*Plan, error) {
	// Build dynamic update query
	query := "UPDATE subscription_plans SET updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')"
	args := []interface{}{planID}
	argCount := 1

	if req.Name != nil {
		argCount++
		query += ", name = $" + string(rune('0'+argCount))
		args = append(args, *req.Name)
	}
	if req.PriceINR != nil {
		argCount++
		query += ", price_inr = $" + string(rune('0'+argCount))
		args = append(args, *req.PriceINR)
	}
	if req.RazorpayPlanID != nil {
		argCount++
		query += ", razorpay_plan_id = $" + string(rune('0'+argCount))
		args = append(args, *req.RazorpayPlanID)
	}
	if req.Features != nil {
		featuresJSON, _ := json.Marshal(*req.Features)
		argCount++
		query += ", features = $" + string(rune('0'+argCount))
		args = append(args, featuresJSON)
	}
	if req.IsActive != nil {
		argCount++
		query += ", is_active = $" + string(rune('0'+argCount))
		args = append(args, *req.IsActive)
	}

	query += " WHERE id = $1 RETURNING id, name, plan_type, price_inr, razorpay_plan_id, features, is_active, created_at, updated_at"

	var p Plan
	var featuresJSON []byte

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.Name, &p.PlanType, &p.PriceINR, &p.RazorpayPlanID,
		&featuresJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(featuresJSON, &p.Features); err != nil {
		p.Features = make(map[string]interface{})
	}

	return &p, nil
}

// Delete soft deletes a plan
func (r *PlansRepository) Delete(ctx context.Context, planID string) error {
	query := `UPDATE subscription_plans SET is_active = false WHERE id = $1`
	result, err := r.db.Exec(ctx, query, planID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrPlanNotFound
	}

	return nil
}
