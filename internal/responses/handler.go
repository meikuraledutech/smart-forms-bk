package responses

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ResponsesHandler struct {
	service *ResponsesService
}

func NewResponsesHandler(service *ResponsesService) *ResponsesHandler {
	return &ResponsesHandler{service: service}
}

// SubmitResponse handles form response submission (public endpoint)
// POST /f/:slug/responses
func (h *ResponsesHandler) SubmitResponse(c *fiber.Ctx) error {
	slug := c.Params("slug")

	var req SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	responseID, err := h.service.SubmitResponse(c.Context(), slug, req)
	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(SubmitResponse{
		Message:    "Response submitted successfully",
		ResponseID: responseID,
	})
}

// GetFormResponses retrieves all responses for a form (protected endpoint)
// GET /forms/:form_id/responses
func (h *ResponsesHandler) GetFormResponses(c *fiber.Ctx) error {
	formID := c.Params("form_id")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	responses, total, err := h.service.GetResponses(c.Context(), formID, limit, offset)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"items":  responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func mapServiceError(err error) error {
	switch err {
	case ErrFormNotFound:
		return fiber.ErrNotFound
	case ErrFormNotPublished:
		return fiber.NewError(fiber.StatusForbidden, "Form is not published")
	case ErrFormNotAccepting:
		return fiber.NewError(fiber.StatusForbidden, "Form is not accepting responses")
	case ErrInvalidInput:
		return fiber.ErrBadRequest
	case ErrInvalidFlowConnection:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid flow connection id")
	default:
		return fiber.ErrInternalServerError
	}
}
