package flows

import (
	"github.com/gofiber/fiber/v2"
)

type FlowHandler struct {
	service *FlowService
}

func NewFlowHandler(service *FlowService) *FlowHandler {
	return &FlowHandler{service: service}
}

func (h *FlowHandler) UpdateFlow(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("form_id")

	var req FlowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	mapping, err := h.service.UpdateFlow(c.Context(), userID, formID, req)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(fiber.Map{
		"message": "Flow updated successfully",
		"mapping": mapping,
	})
}

func (h *FlowHandler) GetFlow(c *fiber.Ctx) error {
	formID := c.Params("form_id")

	connections, err := h.service.GetFlow(c.Context(), formID)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"connections": connections,
		"total":       len(connections),
	})
}

func mapServiceError(err error) error {
	switch err {
	case ErrInvalidInput:
		return fiber.ErrBadRequest
	case ErrNotFound, ErrFormNotFound:
		return fiber.ErrNotFound
	default:
		return fiber.ErrInternalServerError
	}
}
