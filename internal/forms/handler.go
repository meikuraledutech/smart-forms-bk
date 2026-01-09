package forms

import (
	"context"
	"reflect"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// FlowRepository interface for flow operations
type FlowRepository interface {
	GetFlowWithQuestions(ctx context.Context, formID string) ([]map[string]interface{}, error)
	DeleteByFormID(ctx context.Context, formID string) error
	Create(ctx context.Context, formID, questionID string, parentID *string, orderIndex, depthLevel int, isTerminal bool) (interface{}, error)
	CreateQuestion(ctx context.Context, userID, qType, text string) (string, error)
	FindQuestionByText(ctx context.Context, qType, text string) (string, error)
}

// FormsHandler handles HTTP requests
type FormsHandler struct {
	service  *FormsService
	flowRepo FlowRepository
}

// NewFormsHandler creates handler
func NewFormsHandler(service *FormsService) *FormsHandler {
	return &FormsHandler{service: service}
}

// SetFlowRepo sets the flow repository
func (h *FormsHandler) SetFlowRepo(flowRepo FlowRepository) {
	h.flowRepo = flowRepo
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
 TEMPLATE OPERATIONS
========================
*/

// ToggleTemplate toggles is_template flag (super admin only)
// PATCH /admin/forms/:id/template
func (h *FormsHandler) ToggleTemplate(c *fiber.Ctx) error {
	formID := c.Params("id")

	var req struct {
		IsTemplate bool `json:"is_template"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.service.ToggleTemplate(c.Context(), formID, req.IsTemplate)
	if err != nil {
		return mapServiceError(err)
	}

	return c.JSON(fiber.Map{
		"message":     "Template status updated",
		"is_template": req.IsTemplate,
	})
}

// ListTemplates lists all published templates (public)
// GET /templates
func (h *FormsHandler) ListTemplates(c *fiber.Ctx) error {
	// Get templates from cache or database
	templates, err := h.service.ListTemplates(c.Context())
	if err != nil {
		return fiber.ErrInternalServerError
	}

	// Generate ETag based on templates data
	etag := h.service.GenerateTemplatesETag(templates)

	// Check If-None-Match header
	if match := c.Get("If-None-Match"); match == etag {
		return c.SendStatus(fiber.StatusNotModified)
	}

	// Set caching headers - always revalidate with ETag
	// Browser MUST check server on every request but can use cached version if ETag matches
	c.Set("ETag", etag)
	c.Set("Cache-Control", "no-cache") // Always revalidate, but use cached version if ETag matches
	c.Set("Vary", "Accept-Encoding")

	// Set Last-Modified header if templates exist
	if len(templates) > 0 {
		c.Set("Last-Modified", templates[0].UpdatedAt.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}

	return c.JSON(fiber.Map{
		"templates": templates,
		"total":     len(templates),
	})
}

// CloneTemplate clones a template to user's account (authenticated)
// POST /templates/:id/clone
func (h *FormsHandler) CloneTemplate(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	templateID := c.Params("id")

	// 1. Get template metadata
	template, err := h.service.GetTemplateData(c.Context(), templateID)
	if err != nil {
		return mapServiceError(err)
	}

	// 2. Get template flow structure
	flowItems, err := h.flowRepo.GetFlowWithQuestions(c.Context(), templateID)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	// 3. Create new form with template data
	newTitle := "Copy of " + template.Title
	clonedForm, err := h.service.Create(c.Context(), userID, newTitle, template.Description)
	if err != nil {
		return mapServiceError(err)
	}

	// 4. Copy flow structure to new form
	if len(flowItems) > 0 {
		// Build tree from flat flow items
		blocks := h.buildTree(flowItems, nil)

		// Recreate flow in new form
		for i, block := range blocks {
			if err := h.processBlock(c.Context(), userID, clonedForm.ID, block, nil, i, 0); err != nil {
				return fiber.ErrInternalServerError
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Template cloned successfully",
		"form":    clonedForm,
	})
}

// Helper functions for cloning flow
func (h *FormsHandler) buildTree(items []map[string]interface{}, parentID *string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, item := range items {
		itemParentID := item["parent_id"].(*string)

		// Check if this item belongs to current parent
		if (parentID == nil && itemParentID == nil) ||
			(parentID != nil && itemParentID != nil && *parentID == *itemParentID) {

			id := item["id"].(string)
			block := map[string]interface{}{
				"id":       id,
				"type":     item["type"],
				"question": item["question"],
				"children": h.buildTree(items, &id),
			}

			result = append(result, block)
		}
	}

	return result
}

func (h *FormsHandler) processBlock(ctx context.Context, userID, formID string, block map[string]interface{}, parentID *string, orderIndex, depthLevel int) error {
	questionText := block["question"].(string)
	qType := block["type"].(string)

	// Find or create question
	questionID, err := h.flowRepo.FindQuestionByText(ctx, qType, questionText)
	if err != nil {
		// Question doesn't exist, create it
		questionID, err = h.flowRepo.CreateQuestion(ctx, userID, qType, questionText)
		if err != nil {
			return err
		}
	}

	// Get children
	children, _ := block["children"].([]map[string]interface{})
	isTerminal := len(children) == 0

	// Create flow connection
	connection, err := h.flowRepo.Create(ctx, formID, questionID, parentID, orderIndex, depthLevel, isTerminal)
	if err != nil {
		return err
	}

	// Extract ID field using reflection
	val := reflect.ValueOf(connection)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	connID := idField.String()

	// Process children recursively
	for i, child := range children {
		if err := h.processBlock(ctx, userID, formID, child, &connID, i, depthLevel+1); err != nil {
			return err
		}
	}

	return nil
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
