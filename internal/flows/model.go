package flows

import "time"

type FlowConnection struct {
	ID         string    `json:"id"`
	FormID     string    `json:"form_id"`
	QuestionID string    `json:"question_id"`
	ParentID   *string   `json:"parent_id,omitempty"`
	OrderIndex int       `json:"order_index"`
	DepthLevel int       `json:"depth_level"`
	IsTerminal bool      `json:"is_terminal"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"-"`
}

// Request structures
type Block struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Question string   `json:"question"`
	Children []Block  `json:"children"`
}

type FlowRequest struct {
	Blocks []Block `json:"blocks"`
}
