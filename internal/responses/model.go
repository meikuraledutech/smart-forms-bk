package responses

import "time"

// FormResponse represents a submitted response to a form
type FormResponse struct {
	ID              string                 `json:"id"`
	FormID          string                 `json:"form_id"`
	SubmittedAt     time.Time              `json:"submitted_at"`
	TotalTimeSpent  int                    `json:"total_time_spent"`
	FlowPath        []string               `json:"flow_path"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ResponseAnswer represents an answer to a specific question in a response
type ResponseAnswer struct {
	ID               string                 `json:"id"`
	ResponseID       string                 `json:"response_id"`
	FlowConnectionID string                 `json:"flow_connection_id"`
	AnswerText       string                 `json:"answer_text"`
	AnswerValue      map[string]interface{} `json:"answer_value,omitempty"`
	TimeSpent        *int                   `json:"time_spent,omitempty"`
}

// SubmitRequest represents the request body for submitting a response
type SubmitRequest struct {
	Responses []AnswerInput      `json:"responses"`
	Metadata  MetadataInput      `json:"metadata"`
}

// AnswerInput represents a single answer in the submission
type AnswerInput struct {
	FlowConnectionID string                 `json:"flow_connection_id"`
	AnswerText       string                 `json:"answer_text"`
	AnswerValue      map[string]interface{} `json:"answer_value,omitempty"`
	TimeSpent        *int                   `json:"time_spent,omitempty"`
}

// MetadataInput represents the metadata in the submission
type MetadataInput struct {
	TotalTimeSpent int      `json:"total_time_spent"`
	FlowPath       []string `json:"flow_path"`
}

// SubmitResponse represents the response after successful submission
type SubmitResponse struct {
	Message    string `json:"message"`
	ResponseID string `json:"response_id"`
}
