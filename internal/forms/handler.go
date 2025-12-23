package forms

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// FormsHandler handles HTTP requests
type FormsHandler struct {
	service *FormsService
}

// NewFormsHandler creates handler
func NewFormsHandler(service *FormsService) *FormsHandler {
	return &FormsHandler{service: service}
}

/*
========================
 CREATE FORM
POST /forms
========================
*/
func (h *FormsHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	form, err := h.service.Create(
		c.Context(),
		userID,
		req.Title,
		req.Description,
	)

	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(form)
}

/*
========================
 LIST FORMS
GET /forms
========================
*/
func (h *FormsHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	search := c.Query("search", "")

	items, total, err := h.service.List(
		c.Context(),
		userID,
		search,
		limit,
		offset,
	)

	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

/*
========================
 GET FORM BY ID
GET /forms/:id
========================
*/
func (h *FormsHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("id")

	form, err := h.service.GetByID(
		c.Context(),
		userID,
		formID,
	)

	if err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(form)
}

/*
========================
 UPDATE FORM
PATCH /forms/:id
========================
*/
func (h *FormsHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("id")

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.service.Update(
		c.Context(),
		userID,
		formID,
		req.Title,
		req.Description,
		req.Status,
	)

	if err != nil {
		return mapServiceError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

/*
========================
 SOFT DELETE FORM
PATCH /forms/:id/delete
========================
*/
func (h *FormsHandler) SoftDelete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	formID := c.Params("id")

	err := h.service.SoftDelete(
		c.Context(),
		userID,
		formID,
	)

	if err != nil {
		return fiber.ErrNotFound
	}

	return c.SendStatus(fiber.StatusNoContent)
}

/*
========================
 ERROR MAPPING
========================
*/
func mapServiceError(err error) error {
	switch err {
	case ErrInvalidInput:
		return fiber.ErrBadRequest
	case ErrNotFound:
		return fiber.ErrNotFound
	default:
		return fiber.ErrInternalServerError
	}
}
