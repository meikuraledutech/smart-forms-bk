package questions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type QuestionHandler struct {
	service *QuestionService
}

func NewQuestionHandler(service *QuestionService) *QuestionHandler {
	return &QuestionHandler{service: service}
}

func (h *QuestionHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		Type            string         `json:"type"`
		QuestionText    string         `json:"question_text"`
		InputType       string         `json:"input_type"`
		ValidationRules map[string]any `json:"validation_rules"`
		Metadata        map[string]any `json:"metadata"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	question, err := h.service.Create(
		c.Context(),
		userID,
		req.Type,
		req.QuestionText,
		req.InputType,
		req.ValidationRules,
		req.Metadata,
	)

	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(question)
}

func (h *QuestionHandler) List(c *fiber.Ctx) error {
	qType := c.Query("type", "")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	questions, total, err := h.service.List(
		c.Context(),
		qType,
		limit,
		offset,
	)

	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"items":  questions,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *QuestionHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	question, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(question)
}

func (h *QuestionHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Type            string         `json:"type"`
		QuestionText    string         `json:"question_text"`
		InputType       string         `json:"input_type"`
		ValidationRules map[string]any `json:"validation_rules"`
		Metadata        map[string]any `json:"metadata"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.service.Update(
		c.Context(),
		id,
		req.Type,
		req.QuestionText,
		req.InputType,
		req.ValidationRules,
		req.Metadata,
	)

	if err != nil {
		return mapServiceError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *QuestionHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.service.Delete(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func mapServiceError(err error) error {
	switch err {
	case ErrInvalidInput, ErrInvalidType:
		return fiber.ErrBadRequest
	case ErrNotFound:
		return fiber.ErrNotFound
	default:
		return fiber.ErrInternalServerError
	}
}
