package calculators

import (
	"context"
	"time"
)

// NodeCalculator calculates node-level metrics
type NodeCalculator interface {
	Calculate(ctx context.Context, formID string) ([]NodeMetrics, error)
}

// PathCalculator calculates path-level metrics
type PathCalculator interface {
	Calculate(ctx context.Context, formID string) ([]PathMetrics, error)
}

// NodeMetrics represents calculated metrics for a node (matches analytics.NodeMetrics)
type NodeMetrics struct {
	FormID           string
	FlowConnectionID string
	VisitCount       int
	AnswerCount      int
	SkipCount        int
	DropOffCount     int
	TotalTimeSpent   int
	AvgTimeSpent     float64
	CalculatedAt     time.Time
}

// PathMetrics represents calculated metrics for a path
type PathMetrics struct {
	FormID            string
	Path              []string
	OccurrenceCount   int
	AvgCompletionTime float64
	CompletionRate    float64
	CalculatedAt      time.Time
}
