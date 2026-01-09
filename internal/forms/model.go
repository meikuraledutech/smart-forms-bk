package forms

import "time"

// Form represents a form metadata entity
type Form struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"-"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Status             string     `json:"status"`
	AutoSlug           *string    `json:"auto_slug,omitempty"`
	CustomSlug         *string    `json:"custom_slug,omitempty"`
	AcceptingResponses bool       `json:"accepting_responses"`
	PublishedAt        *time.Time `json:"published_at,omitempty"`
	IsTemplate         bool       `json:"is_template"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"-"`
}
