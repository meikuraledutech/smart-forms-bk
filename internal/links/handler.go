package links

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

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

	// Generate ETag based on form content
	etag := generateETag(form)

	// Check If-None-Match header for conditional request
	clientETag := c.Get("If-None-Match")
	if clientETag == etag {
		// Content hasn't changed, return 304 Not Modified
		c.Set("ETag", etag)
		return c.SendStatus(fiber.StatusNotModified)
	}

	// Set cache headers
	// no-cache: Browser must revalidate with server (but can use ETag for 304)
	// s-maxage=600: CDN (Vercel) caches for 10 minutes
	// This ensures form updates are immediately visible while still benefiting from CDN
	c.Set("Cache-Control", "no-cache, s-maxage=600")
	c.Set("ETag", etag)

	return c.JSON(form)
}

// generateETag creates a unique hash of the form content
func generateETag(form *PublicForm) string {
	// Marshal form to JSON for consistent hashing
	data, err := json.Marshal(form)
	if err != nil {
		// Fallback to form ID if marshaling fails
		return fmt.Sprintf(`"%s"`, form.ID)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(data)
	// Return first 16 hex characters wrapped in quotes (standard ETag format)
	return fmt.Sprintf(`"%.16x"`, hash[:8])
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
