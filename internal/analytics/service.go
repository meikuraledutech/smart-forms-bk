package analytics

import (
	"context"
	"smart-forms/internal/analytics/calculators"
)

type AnalyticsService struct {
	repo            *AnalyticsRepository
	nodeCalculator  calculators.NodeCalculator
	pathCalculator  calculators.PathCalculator
}

func NewAnalyticsService(repo *AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		repo:           repo,
		nodeCalculator: calculators.NewNodeCalculator(repo),
		// TODO: Initialize path calculator when implementing path analytics
		// pathCalculator:  calculators.NewPathCalculator(repo),
	}
}

// GetAnalytics retrieves or calculates analytics for a form
func (s *AnalyticsService) GetAnalytics(ctx context.Context, formID, userID string) (*AnalyticsOverview, error) {
	// TODO: Implement
	// 1. Check if analytics exist and are fresh
	// 2. If not, trigger calculation
	// 3. Return analytics

	return nil, ErrCalculationPending
}

// GetStatus returns the current calculation status
func (s *AnalyticsService) GetStatus(ctx context.Context, formID string) (*StatusResponse, error) {
	status, err := s.repo.GetStatus(ctx, formID)
	if err != nil {
		// No status record means analytics haven't been calculated yet
		return &StatusResponse{
			Status:  "not_started",
			Message: "Analytics have not been calculated yet",
		}, nil
	}

	response := &StatusResponse{
		Status:       status.Status,
		CalculatedAt: status.CalculatedAt,
	}

	// Add helpful messages based on status
	switch status.Status {
	case "calculating":
		response.Message = "Analytics calculation in progress"
	case "completed":
		response.Message = "Analytics are ready"
	case "failed":
		response.Message = "Analytics calculation failed"
	case "pending":
		response.Message = "Analytics calculation queued"
	}

	return response, nil
}

// RefreshAnalytics triggers recalculation
func (s *AnalyticsService) RefreshAnalytics(ctx context.Context, formID, userID string) error {
	// TODO: Implement
	// 1. Mark as stale/calculating
	// 2. Start calculation
	// 3. Update status when done

	return nil
}

// CalculateAnalytics performs the actual calculation
func (s *AnalyticsService) CalculateAnalytics(ctx context.Context, formID string) error {
	// TODO: Implement
	// 1. Use calculators to compute metrics
	// 2. Store in analytics tables
	// 3. Update status to completed

	// Example flow:
	// nodeMetrics := s.nodeCalculator.Calculate(ctx, formID)
	// pathMetrics := s.pathCalculator.Calculate(ctx, formID)
	// s.repo.SaveNodeMetrics(ctx, nodeMetrics)
	// s.repo.SavePathMetrics(ctx, pathMetrics)

	return nil
}

// GetNodeMetrics retrieves node-level analytics
func (s *AnalyticsService) GetNodeMetrics(ctx context.Context, formID string) ([]NodeMetrics, error) {
	// Check if metrics already exist
	existing, err := s.repo.GetNodeMetrics(ctx, formID)
	if err == nil && len(existing) > 0 {
		// Return existing metrics
		return existing, nil
	}

	// Metrics don't exist, calculate them
	calcMetrics, err := s.nodeCalculator.Calculate(ctx, formID)
	if err != nil {
		return nil, err
	}

	if len(calcMetrics) == 0 {
		return nil, ErrNoResponses
	}

	// Convert calculator metrics to analytics metrics
	metrics := make([]NodeMetrics, len(calcMetrics))
	for i, cm := range calcMetrics {
		metrics[i] = NodeMetrics{
			FormID:           cm.FormID,
			FlowConnectionID: cm.FlowConnectionID,
			VisitCount:       cm.VisitCount,
			AnswerCount:      cm.AnswerCount,
			SkipCount:        cm.SkipCount,
			DropOffCount:     cm.DropOffCount,
			TotalTimeSpent:   cm.TotalTimeSpent,
			AvgTimeSpent:     cm.AvgTimeSpent,
			CalculatedAt:     cm.CalculatedAt,
		}
	}

	// Save the calculated metrics
	if err := s.repo.SaveNodeMetrics(ctx, metrics); err != nil {
		return nil, err
	}

	// Update status to completed
	if err := s.repo.UpdateStatus(ctx, formID, "completed"); err != nil {
		// Non-fatal, just log or ignore
	}

	// Read back from DB to get enriched data (with question text and type)
	enrichedMetrics, err := s.repo.GetNodeMetrics(ctx, formID)
	if err != nil {
		// Fallback to calculated metrics if read fails
		return metrics, nil
	}

	return enrichedMetrics, nil
}

// GetPathMetrics retrieves path-level analytics
func (s *AnalyticsService) GetPathMetrics(ctx context.Context, formID string) ([]PathMetrics, error) {
	// TODO: Implement
	return nil, nil
}
