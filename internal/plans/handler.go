package plans

import (
	"github.com/gofiber/fiber/v2"
)

type PlansHandler struct {
	service *PlansService
}

func NewPlansHandler(service *PlansService) *PlansHandler {
	return &PlansHandler{service: service}
}

// ListActivePlans retrieves all active plans (public endpoint)
// GET /plans
func (h *PlansHandler) ListActivePlans(c *fiber.Ctx) error {
	plans, err := h.service.ListPlans(c.Context(), true)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"plans": plans,
	})
}

// ListAllPlans retrieves all plans including inactive (super admin only)
// GET /admin/plans
func (h *PlansHandler) ListAllPlans(c *fiber.Ctx) error {
	plans, err := h.service.ListPlans(c.Context(), false)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"plans": plans,
	})
}

// GetPlan retrieves a single plan
// GET /admin/plans/:id
func (h *PlansHandler) GetPlan(c *fiber.Ctx) error {
	planID := c.Params("id")

	plan, err := h.service.GetPlan(c.Context(), planID)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(plan)
}

// CreatePlan creates a new plan (super admin only)
// POST /admin/plans
func (h *PlansHandler) CreatePlan(c *fiber.Ctx) error {
	var req CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	plan, err := h.service.CreatePlan(c.Context(), req)
	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(plan)
}

// UpdatePlan updates a plan (super admin only)
// PATCH /admin/plans/:id
func (h *PlansHandler) UpdatePlan(c *fiber.Ctx) error {
	planID := c.Params("id")

	var req UpdatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	plan, err := h.service.UpdatePlan(c.Context(), planID, req)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(plan)
}

// DeletePlan soft deletes a plan (super admin only)
// DELETE /admin/plans/:id
func (h *PlansHandler) DeletePlan(c *fiber.Ctx) error {
	planID := c.Params("id")

	err := h.service.DeletePlan(c.Context(), planID)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(fiber.Map{
		"message": "Plan deleted successfully",
	})
}

func mapServiceError(err error) error {
	switch err {
	case ErrPlanNotFound:
		return fiber.ErrNotFound
	case ErrPlanAlreadyExists:
		return fiber.NewError(fiber.StatusConflict, "Plan with this name already exists")
	case ErrInvalidInput:
		return fiber.ErrBadRequest
	default:
		return fiber.ErrInternalServerError
	}
}
