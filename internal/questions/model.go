package questions

import "time"

type Question struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	QuestionText    string                 `json:"question_text"`
	InputType       *string                `json:"input_type,omitempty"`
	ValidationRules map[string]interface{} `json:"validation_rules,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy       *string                `json:"created_by,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       *time.Time             `json:"-"`
}
