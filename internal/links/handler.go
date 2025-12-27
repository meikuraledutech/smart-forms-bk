package links

import (
	"github.com/gofiber/fiber/v2"
)

type LinksHandler struct {
	service *LinksService
}

func NewLinksHandler(service *LinksService) *LinksHandler {
	return &LinksHandler{service: service}
}

// PublishForm handles publishing a form
// PATCH /forms/:form_id/publish
func (h *LinksHandler) PublishForm(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("form_id")

	var req PublishRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	var customSlug *string
	if req.CustomSlug != "" {
		customSlug = &req.CustomSlug
	}

	autoSlug, customSlugResult, err := h.service.PublishForm(c.Context(), formID, userID, customSlug)
	if err != nil {
		return mapServiceError(err)
	}

	response := PublishResponse{
		AutoSlug: autoSlug,
		AutoURL:  "/f/" + autoSlug,
	}

	if customSlugResult != nil {
		response.CustomSlug = customSlugResult
		customURL := "/f/" + *customSlugResult
		response.CustomURL = &customURL
	}

	return c.JSON(fiber.Map{
		"message": "Form published successfully",
		"links":   response,
	})
}

// ToggleAcceptingResponses handles toggling accepting responses
// PATCH /forms/:form_id/accepting-responses
func (h *LinksHandler) ToggleAcceptingResponses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("form_id")

	var req ToggleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.service.ToggleAcceptingResponses(c.Context(), formID, userID, req.Accepting)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"message":            "Updated successfully",
		"accepting_responses": req.Accepting,
	})
}

// GetPublicForm handles getting a public form by slug
// GET /f/:slug
func (h *LinksHandler) GetPublicForm(c *fiber.Ctx) error {
	slug := c.Params("slug")

	form, err := h.service.GetPublicForm(c.Context(), slug)
	if err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(form)
}

func mapServiceError(err error) error {
	switch err {
	case ErrFormNotFound:
		return fiber.ErrNotFound
	case ErrSlugTaken:
		return fiber.NewError(fiber.StatusConflict, "Slug already taken")
	case ErrInvalidSlug:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid slug format (3-50 chars, lowercase, numbers, hyphens only)")
	case ErrInvalidInput:
		return fiber.ErrBadRequest
	default:
		return fiber.ErrInternalServerError
	}
}
