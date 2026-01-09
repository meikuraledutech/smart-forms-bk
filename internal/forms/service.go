package forms

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"smart-forms/internal/cache"
)

// Allowed form statuses (v1)
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

// Service-level errors
var (
	ErrInvalidInput = errors.New("invalid input")
)

// FormsService coordinates business logic
type FormsService struct {
	repo  *FormsRepository
	cache *cache.Cache
}

// NewFormsService creates service
func NewFormsService(repo *FormsRepository, cacheClient *cache.Cache) *FormsService {
	return &FormsService{
		repo:  repo,
		cache: cacheClient,
	}
}

/*
========================
 CREATE FORM
========================
*/
func (s *FormsService) Create(
	ctx context.Context,
	userID string,
	title string,
	description string,
) (*Form, error) {

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" {
		return nil, ErrInvalidInput
	}

	return s.repo.Create(ctx, userID, title, description)
}

/*
========================
 GET FORM
========================
*/
func (s *FormsService) GetByID(
	ctx context.Context,
	userID string,
	formID string,
) (*Form, error) {

	return s.repo.GetByID(ctx, userID, formID)
}

/*
========================
 LIST FORMS
========================
*/
func (s *FormsService) List(
	ctx context.Context,
	userID string,
	search string,
	limit int,
	offset int,
) ([]Form, int, error) {

	search = strings.TrimSpace(search)

	return s.repo.List(ctx, userID, search, limit, offset)
}

/*
========================
 UPDATE FORM (PATCH)
========================
*/
func (s *FormsService) Update(
	ctx context.Context,
	userID string,
	formID string,
	title string,
	description string,
	status string,
) error {

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" {
		return ErrInvalidInput
	}

	if !isValidStatus(status) {
		return ErrInvalidInput
	}

	// Get slugs before update (for cache invalidation)
	autoSlug, customSlug, _ := s.repo.GetFormSlugs(ctx, formID)

	// Check if this is a template before update
	existingForm, _ := s.repo.GetByID(ctx, userID, formID)
	isTemplate := existingForm != nil && existingForm.IsTemplate

	// Update in database
	err := s.repo.Update(ctx, userID, formID, title, description, status)
	if err != nil {
		return err
	}

	// Invalidate cache after successful update
	// Delete by form ID
	s.cache.Delete(cache.FormIDKey(formID))

	// Delete by slugs (if form was published)
	if autoSlug != nil && *autoSlug != "" {
		s.cache.Delete(cache.FormSlugKey(*autoSlug))
	}
	if customSlug != nil && *customSlug != "" {
		s.cache.Delete(cache.FormSlugKey(*customSlug))
	}

	// If this is a template, invalidate templates list cache
	if isTemplate {
		s.InvalidateTemplatesCache()
	}

	return nil
}

/*
========================
 SOFT DELETE
========================
*/
func (s *FormsService) SoftDelete(
	ctx context.Context,
	userID string,
	formID string,
) error {

	// Get slugs before delete (for cache invalidation)
	autoSlug, customSlug, _ := s.repo.GetFormSlugs(ctx, formID)

	// Check if this is a template before delete
	existingForm, _ := s.repo.GetByID(ctx, userID, formID)
	isTemplate := existingForm != nil && existingForm.IsTemplate

	// Delete from database
	err := s.repo.SoftDelete(ctx, userID, formID)
	if err != nil {
		return err
	}

	// Invalidate cache after successful delete
	// Delete by form ID
	s.cache.Delete(cache.FormIDKey(formID))

	// Delete by slugs (if form was published)
	if autoSlug != nil && *autoSlug != "" {
		s.cache.Delete(cache.FormSlugKey(*autoSlug))
	}
	if customSlug != nil && *customSlug != "" {
		s.cache.Delete(cache.FormSlugKey(*customSlug))
	}

	// If this was a template, invalidate templates list cache
	if isTemplate {
		s.InvalidateTemplatesCache()
	}

	return nil
}

/*
========================
 HELPERS
========================
*/
func isValidStatus(status string) bool {
	switch status {
	case StatusDraft, StatusPublished:
		return true
	default:
		return false
	}
}

/*
========================
 TEMPLATE OPERATIONS
========================
*/

// ToggleTemplate toggles is_template flag (super admin only)
func (s *FormsService) ToggleTemplate(
	ctx context.Context,
	formID string,
	isTemplate bool,
) error {
	err := s.repo.ToggleTemplate(ctx, formID, isTemplate)
	if err != nil {
		return err
	}

	// Invalidate templates cache
	s.InvalidateTemplatesCache()

	return nil
}

// ListTemplates lists all published templates (public)
func (s *FormsService) ListTemplates(
	ctx context.Context,
) ([]Form, error) {
	// Try cache first
	cacheKey := "templates:list"
	if cached, found := s.cache.Get(cacheKey); found {
		if templates, ok := cached.([]Form); ok {
			return templates, nil
		}
	}

	// Cache miss - fetch from database
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 7 days
	s.cache.Set(cacheKey, templates, 7*24*time.Hour)

	return templates, nil
}

// GenerateTemplatesETag generates an ETag for templates list
func (s *FormsService) GenerateTemplatesETag(templates []Form) string {
	// Create a hash based on template count, IDs and updated_at timestamps
	// This ensures ETag changes whenever templates are added/removed/modified
	type etagData struct {
		Count     int
		Templates []Form
	}

	data, _ := json.Marshal(etagData{
		Count:     len(templates),
		Templates: templates,
	})
	hash := sha256.Sum256(data)
	return fmt.Sprintf(`W/"%x"`, hash[:16]) // Weak ETag with first 16 bytes for better uniqueness
}

// InvalidateTemplatesCache invalidates the templates list cache
func (s *FormsService) InvalidateTemplatesCache() {
	s.cache.Delete("templates:list")
}

// GetTemplateData gets template data for cloning
func (s *FormsService) GetTemplateData(
	ctx context.Context,
	templateID string,
) (*Form, error) {
	return s.repo.GetTemplateWithFlow(ctx, templateID)
}
