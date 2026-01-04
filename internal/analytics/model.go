package analytics

import "time"

// AnalyticsStatus represents the calculation status for a form
type AnalyticsStatus struct {
	FormID        string    `json:"form_id"`
	Status        string    `json:"status"` // pending, calculating, completed, failed
	CalculatedAt  *time.Time `json:"calculated_at,omitempty"`
	TriggeredBy   string    `json:"triggered_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NodeMetrics represents analytics for a single node (flow_connection)
type NodeMetrics struct {
	FormID           string    `json:"form_id"`
	FlowConnectionID string    `json:"flow_connection_id"`
	QuestionText     string    `json:"question_text"`
	QuestionType     string    `json:"question_type"`
	VisitCount       int       `json:"visit_count"`
	AnswerCount      int       `json:"answer_count"`
	SkipCount        int       `json:"skip_count"`
	DropOffCount     int       `json:"drop_off_count"`
	TotalTimeSpent   int       `json:"total_time_spent"`
	AvgTimeSpent     float64   `json:"avg_time_spent"`
	CalculatedAt     time.Time `json:"calculated_at"`
}

// PathMetrics represents analytics for a specific path through the form
type PathMetrics struct {
	FormID            string    `json:"form_id"`
	Path              []string  `json:"path"`
	OccurrenceCount   int       `json:"occurrence_count"`
	AvgCompletionTime float64   `json:"avg_completion_time"`
	CompletionRate    float64   `json:"completion_rate"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// AnalyticsOverview represents the complete analytics for a form
type AnalyticsOverview struct {
	FormID           string        `json:"form_id"`
	TotalResponses   int           `json:"total_responses"`
	AvgCompletionTime float64      `json:"avg_completion_time"`
	CompletionRate   float64       `json:"completion_rate"`
	NodeMetrics      []NodeMetrics `json:"node_metrics"`
	TopPaths         []PathMetrics `json:"top_paths"`
	CalculatedAt     time.Time     `json:"calculated_at"`
}

// StatusResponse for polling endpoint
type StatusResponse struct {
	Status       string     `json:"status"`
	Progress     int        `json:"progress,omitempty"`
	CalculatedAt *time.Time `json:"calculated_at,omitempty"`
	Message      string     `json:"message,omitempty"`
}

// FlowTransition represents a transition from one node to another
type FlowTransition struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

// FlowAnalytics represents the complete flow analytics for Sankey visualization
type FlowAnalytics struct {
	FormID  string           `json:"form_id"`
	Flows   []FlowTransition `json:"flows"`
	Mermaid string           `json:"mermaid"`
}
