package plans

import (
	"context"
	"strings"
)

type PlansService struct {
	repo *PlansRepository
}

func NewPlansService(repo *PlansRepository) *PlansService {
	return &PlansService{repo: repo}
}

// ListPlans retrieves all plans (for users, show only active)
func (s *PlansService) ListPlans(ctx context.Context, activeOnly bool) ([]Plan, error) {
	return s.repo.List(ctx, activeOnly)
}

// GetPlan retrieves a plan by ID
func (s *PlansService) GetPlan(ctx context.Context, planID string) (*Plan, error) {
	return s.repo.GetByID(ctx, planID)
}

// CreatePlan creates a new plan (super admin only)
func (s *PlansService) CreatePlan(ctx context.Context, req CreatePlanRequest) (*Plan, error) {
	// Validate input
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.PlanType == "" {
		return nil, ErrInvalidInput
	}

	// Validate plan type
	if req.PlanType != "free" && req.PlanType != "monthly" && req.PlanType != "yearly" {
		return nil, ErrInvalidInput
	}

	// Ensure features are set
	if req.Features == nil {
		req.Features = make(map[string]interface{})
	}

	// Set default features based on plan type
	if req.PlanType == "free" {
		req.Features["data_retention_days"] = 7
		req.Features["can_export"] = false
		req.Features["full_analytics"] = false
	} else {
		// Pro plans (monthly/yearly)
		req.Features["data_retention_days"] = 0 // unlimited
		req.Features["can_export"] = true
		req.Features["full_analytics"] = true
	}

	return s.repo.Create(ctx, req)
}

// UpdatePlan updates a plan (super admin only)
func (s *PlansService) UpdatePlan(ctx context.Context, planID string, req UpdatePlanRequest) (*Plan, error) {
	// Validate name if provided
	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if *req.Name == "" {
			return nil, ErrInvalidInput
		}
	}

	return s.repo.Update(ctx, planID, req)
}

// DeletePlan soft deletes a plan (super admin only)
func (s *PlansService) DeletePlan(ctx context.Context, planID string) error {
	return s.repo.Delete(ctx, planID)
}
