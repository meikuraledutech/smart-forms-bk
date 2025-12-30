package calculators

import (
	"context"
)

type nodeCalculator struct {
	repo Repository
}

// Repository interface for data access
type Repository interface {
	GetResponseData(ctx context.Context, formID string) ([]ResponseData, error)
}

// ResponseData represents raw response data for calculations
type ResponseData struct {
	ResponseID     string
	FlowPath       []string
	TotalTimeSpent int
	Answers        []AnswerData
}

// AnswerData represents raw answer data
type AnswerData struct {
	FlowConnectionID string
	AnswerText       string
	TimeSpent        *int
}

func NewNodeCalculator(repo Repository) NodeCalculator {
	return &nodeCalculator{repo: repo}
}

// Calculate computes node-level metrics
func (c *nodeCalculator) Calculate(ctx context.Context, formID string) ([]NodeMetrics, error) {
	// Get all response data
	responses, err := c.repo.GetResponseData(ctx, formID)
	if err != nil {
		return nil, err
	}

	if len(responses) == 0 {
		return []NodeMetrics{}, nil
	}

	// Map to store metrics per node
	nodeMetrics := make(map[string]*NodeMetrics)

	// Process each response
	for _, response := range responses {
		// Track which nodes were answered
		answeredNodes := make(map[string]bool)
		answerTimeMap := make(map[string]int)

		for _, answer := range response.Answers {
			answeredNodes[answer.FlowConnectionID] = true
			if answer.TimeSpent != nil {
				answerTimeMap[answer.FlowConnectionID] = *answer.TimeSpent
			}
		}

		// Process flow path
		for i, nodeID := range response.FlowPath {
			// Initialize node if not exists
			if _, exists := nodeMetrics[nodeID]; !exists {
				nodeMetrics[nodeID] = &NodeMetrics{
					FormID:           formID,
					FlowConnectionID: nodeID,
				}
			}

			node := nodeMetrics[nodeID]

			// Increment visit count
			node.VisitCount++

			// Check if answered
			if answeredNodes[nodeID] {
				node.AnswerCount++

				// Add time spent
				if timeSpent, hasTime := answerTimeMap[nodeID]; hasTime {
					node.TotalTimeSpent += timeSpent
				}
			} else {
				// Visited but not answered = skip
				node.SkipCount++
			}

			// Check if this is the last node in path (drop-off)
			if i == len(response.FlowPath)-1 {
				node.DropOffCount++
			}
		}
	}

	// Calculate averages and convert to slice
	var results []NodeMetrics
	for _, node := range nodeMetrics {
		if node.AnswerCount > 0 {
			node.AvgTimeSpent = float64(node.TotalTimeSpent) / float64(node.AnswerCount)
		}
		results = append(results, *node)
	}

	return results, nil
}
