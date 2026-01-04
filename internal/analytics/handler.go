package analytics

import (
	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	service *AnalyticsService
}

func NewAnalyticsHandler(service *AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// GetAnalytics retrieves analytics for a form
// GET /forms/:form_id/analytics
func (h *AnalyticsHandler) GetAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("form_id")

	// TODO: Implement analytics retrieval
	// Check if analytics exist and are fresh
	// If not, trigger calculation
	// Return analytics or status

	return c.JSON(fiber.Map{
		"message": "Analytics endpoint - to be implemented",
		"form_id": formID,
		"user_id": userID,
	})
}

// GetAnalyticsStatus checks the status of analytics calculation
// GET /forms/:form_id/analytics/status
func (h *AnalyticsHandler) GetAnalyticsStatus(c *fiber.Ctx) error {
	formID := c.Params("form_id")

	status, err := h.service.GetStatus(c.Context(), formID)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(status)
}

// RefreshAnalytics triggers recalculation of analytics
// POST /forms/:form_id/analytics/refresh
func (h *AnalyticsHandler) RefreshAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("form_id")

	// TODO: Implement refresh logic
	// Mark analytics as stale
	// Trigger calculation
	// Return status

	return c.JSON(fiber.Map{
		"message":   "Calculation triggered",
		"form_id":   formID,
		"user_id":   userID,
		"status":    "calculating",
	})
}

// GetNodeAnalytics retrieves node-specific analytics
// GET /forms/:form_id/analytics/nodes
func (h *AnalyticsHandler) GetNodeAnalytics(c *fiber.Ctx) error {
	formID := c.Params("form_id")

	metrics, err := h.service.GetNodeMetrics(c.Context(), formID)
	if err != nil {
		return mapServiceError(err)
	}

	// Calculate form-level summary
	totalResponses := 0
	totalAnswers := 0
	var totalAvgTime float64

	if len(metrics) > 0 {
		// Total responses = visit count of first node (all users start there)
		totalResponses = metrics[0].VisitCount

		// Sum up all answers across nodes
		for _, m := range metrics {
			totalAnswers += m.AnswerCount
			totalAvgTime += m.AvgTimeSpent
		}

		// Average time per node
		if len(metrics) > 0 {
			totalAvgTime = totalAvgTime / float64(len(metrics))
		}
	}

	return c.JSON(fiber.Map{
		"form_id": formID,
		"summary": fiber.Map{
			"total_responses": totalResponses,
			"total_answers":   totalAnswers,
			"avg_time_per_node": totalAvgTime,
			"total_nodes":     len(metrics),
		},
		"nodes": metrics,
	})
}

// GetPathAnalytics retrieves path-specific analytics
// GET /forms/:form_id/analytics/paths
func (h *AnalyticsHandler) GetPathAnalytics(c *fiber.Ctx) error {
	formID := c.Params("form_id")

	// TODO: Implement path analytics retrieval

	return c.JSON(fiber.Map{
		"message": "Path analytics endpoint - to be implemented",
		"form_id": formID,
	})
}

// GetFlowAnalytics retrieves flow transitions for Sankey diagram
// GET /forms/:form_id/analytics/flow
func (h *AnalyticsHandler) GetFlowAnalytics(c *fiber.Ctx) error {
	formID := c.Params("form_id")

	flowAnalytics, err := h.service.GetFlowAnalytics(c.Context(), formID)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(flowAnalytics)
}

func mapServiceError(err error) error {
	switch err {
	case ErrFormNotFound:
		return fiber.ErrNotFound
	case ErrNoResponses:
		return fiber.NewError(fiber.StatusNotFound, "No responses found for this form")
	case ErrCalculationFailed:
		return fiber.ErrInternalServerError
	case ErrCalculationPending:
		return fiber.NewError(fiber.StatusAccepted, "Analytics calculation is in progress")
	case ErrInvalidInput:
		return fiber.ErrBadRequest
	default:
		return fiber.ErrInternalServerError
	}
}
