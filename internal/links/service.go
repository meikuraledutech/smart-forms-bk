package links

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"regexp"
	"strings"
)

type LinksService struct {
	repo *LinksRepository
}

func NewLinksService(repo *LinksRepository) *LinksService {
	return &LinksService{repo: repo}
}

// PublishForm publishes a form with auto-generated and optional custom slug
func (s *LinksService) PublishForm(ctx context.Context, formID, userID string, customSlug *string) (string, *string, error) {
	// Generate auto slug
	autoSlug := s.generateAutoSlug()

	// Validate and process custom slug if provided
	if customSlug != nil && *customSlug != "" {
		*customSlug = strings.TrimSpace(*customSlug)

		// Validate custom slug format
		if !s.isValidSlug(*customSlug) {
			return "", nil, ErrInvalidSlug
		}

		// Check if custom slug is taken
		exists, err := s.repo.CheckSlugExists(ctx, *customSlug)
		if err != nil {
			return "", nil, err
		}
		if exists {
			return "", nil, ErrSlugTaken
		}
	} else {
		customSlug = nil
	}

	// Publish the form
	err := s.repo.PublishForm(ctx, formID, userID, autoSlug, customSlug)
	if err != nil {
		return "", nil, err
	}

	return autoSlug, customSlug, nil
}

// ToggleAcceptingResponses toggles whether a form accepts responses
func (s *LinksService) ToggleAcceptingResponses(ctx context.Context, formID, userID string, accepting bool) error {
	return s.repo.ToggleAcceptingResponses(ctx, formID, userID, accepting)
}

// GetPublicForm retrieves a form by slug for public view
func (s *LinksService) GetPublicForm(ctx context.Context, slug string) (*PublicForm, error) {
	formID, title, description, acceptingResponses, err := s.repo.GetFormBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Get flow structure
	items, err := s.repo.GetFlowForPublicForm(ctx, formID)
	if err != nil {
		return nil, err
	}

	// Build tree structure
	blocks := s.buildTree(items, nil)

	return &PublicForm{
		ID:                 formID,
		Title:              title,
		Description:        description,
		AcceptingResponses: acceptingResponses,
		Flow: map[string]interface{}{
			"blocks": blocks,
		},
	}, nil
}

// buildTree recursively builds nested block structure
func (s *LinksService) buildTree(items []map[string]interface{}, parentID *string) []map[string]interface{} {
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
				"children": s.buildTree(items, &id),
			}

			result = append(result, block)
		}
	}

	return result
}

// GetFormSlugs retrieves the slugs for a form
func (s *LinksService) GetFormSlugs(ctx context.Context, formID, userID string) (string, *string, error) {
	return s.repo.GetFormSlugs(ctx, formID, userID)
}

// generateAutoSlug generates a random slug
func (s *LinksService) generateAutoSlug() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:11]
}

// isValidSlug validates slug format (alphanumeric and hyphens only)
func (s *LinksService) isValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, slug)
	return matched
}
