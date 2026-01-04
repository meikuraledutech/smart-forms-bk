package calculators

import (
	"context"
)

type flowCalculator struct {
	repo Repository
}

func NewFlowCalculator(repo Repository) FlowCalculator {
	return &flowCalculator{repo: repo}
}

// Calculate computes flow transitions for Sankey diagram
func (c *flowCalculator) Calculate(ctx context.Context, formID string) ([]FlowTransition, error) {
	// Get all response data
	responses, err := c.repo.GetResponseData(ctx, formID)
	if err != nil {
		return nil, err
	}

	if len(responses) == 0 {
		return []FlowTransition{}, nil
	}

	// Map to track transitions: "sourceID->targetID" -> count
	transitionCounts := make(map[string]*struct {
		sourceID  string
		targetID  string
		count     int
		isDropOff bool
	})

	// Process each response
	for _, response := range responses {
		// Process transitions in flow path
		for i := 0; i < len(response.FlowPath); i++ {
			currentNodeID := response.FlowPath[i]

			if i < len(response.FlowPath)-1 {
				// Transition to next node
				nextNodeID := response.FlowPath[i+1]
				key := currentNodeID + "->" + nextNodeID

				if _, exists := transitionCounts[key]; !exists {
					transitionCounts[key] = &struct {
						sourceID  string
						targetID  string
						count     int
						isDropOff bool
					}{sourceID: currentNodeID, targetID: nextNodeID, count: 0, isDropOff: false}
				}
				transitionCounts[key].count++
			} else {
				// Last node - transition to drop-off
				key := currentNodeID + "->DROP_OFF"

				if _, exists := transitionCounts[key]; !exists {
					transitionCounts[key] = &struct {
						sourceID  string
						targetID  string
						count     int
						isDropOff bool
					}{sourceID: currentNodeID, targetID: "DROP_OFF", count: 0, isDropOff: true}
				}
				transitionCounts[key].count++
			}
		}
	}

	// Convert to FlowTransition array
	flows := make([]FlowTransition, 0, len(transitionCounts))
	for _, transition := range transitionCounts {
		flows = append(flows, FlowTransition{
			SourceID:  transition.sourceID,
			TargetID:  transition.targetID,
			Value:     transition.count,
			IsDropOff: transition.isDropOff,
			// Source and Target text will be enriched by service/repository layer
		})
	}

	return flows, nil
}
