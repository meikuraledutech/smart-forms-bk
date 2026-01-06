package plans

import (
	"errors"
	"time"
)

// Plan represents a subscription plan
type Plan struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	PlanType       string                 `json:"plan_type"` // 'free', 'monthly', 'yearly'
	PriceINR       int                    `json:"price_inr"` // Price in paise
	RazorpayPlanID *string                `json:"razorpay_plan_id,omitempty"`
	Features       map[string]interface{} `json:"features"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreatePlanRequest for creating a new plan
type CreatePlanRequest struct {
	Name           string                 `json:"name"`
	PlanType       string                 `json:"plan_type"`
	PriceINR       int                    `json:"price_inr"`
	RazorpayPlanID *string                `json:"razorpay_plan_id,omitempty"`
	Features       map[string]interface{} `json:"features"`
}

// UpdatePlanRequest for updating a plan
type UpdatePlanRequest struct {
	Name           *string                 `json:"name,omitempty"`
	PriceINR       *int                    `json:"price_inr,omitempty"`
	RazorpayPlanID *string                 `json:"razorpay_plan_id,omitempty"`
	Features       *map[string]interface{} `json:"features,omitempty"`
	IsActive       *bool                   `json:"is_active,omitempty"`
}

// Errors
var (
	ErrPlanNotFound      = errors.New("plan not found")
	ErrPlanAlreadyExists = errors.New("plan with this name already exists")
	ErrInvalidInput      = errors.New("invalid input")
)
