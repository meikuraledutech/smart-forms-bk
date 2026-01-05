package forms

import (
	"context"
	"errors"
	"strings"

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

	// Update in database
	err := s.repo.Update(ctx, userID, formID, title, description, status)
	if err != nil {
		return err
	}

	// Invalidate cache after successful update
	cacheKey := cache.FormIDKey(formID)
	s.cache.Delete(cacheKey)

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

	// Delete from database
	err := s.repo.SoftDelete(ctx, userID, formID)
	if err != nil {
		return err
	}

	// Invalidate cache after successful delete
	cacheKey := cache.FormIDKey(formID)
	s.cache.Delete(cacheKey)

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
