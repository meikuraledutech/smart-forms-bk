package flows

import (
	"context"
	"strings"

	"smart-forms/internal/cache"
)

type FlowService struct {
	repo  *FlowRepository
	cache *cache.Cache
}

func NewFlowService(repo *FlowRepository, cacheClient *cache.Cache) *FlowService {
	return &FlowService{
		repo:  repo,
		cache: cacheClient,
	}
}

func (s *FlowService) UpdateFlow(ctx context.Context, userID, formID string, req FlowRequest) (map[string]string, error) {
	if len(req.Blocks) == 0 {
		return nil, ErrInvalidInput
	}

	// Verify user owns the form
	if err := s.repo.VerifyFormOwnership(ctx, formID, userID); err != nil {
		return nil, err
	}

	// Soft delete existing flow
	if err := s.repo.DeleteByFormID(ctx, formID); err != nil {
		return nil, err
	}

	// Process blocks recursively and collect ID mapping
	mapping := make(map[string]string)
	for i, block := range req.Blocks {
		if err := s.processBlock(ctx, userID, formID, block, nil, i, 0, mapping); err != nil {
			return nil, err
		}
	}

	// Invalidate cache after flow structure changes
	// Delete by form ID
	s.cache.Delete(cache.FormIDKey(formID))

	// Delete by slugs (if form was published)
	autoSlug, customSlug, _ := s.repo.GetFormSlugs(ctx, formID)
	if autoSlug != nil && *autoSlug != "" {
		s.cache.Delete(cache.FormSlugKey(*autoSlug))
	}
	if customSlug != nil && *customSlug != "" {
		s.cache.Delete(cache.FormSlugKey(*customSlug))
	}

	return mapping, nil
}

func (s *FlowService) processBlock(ctx context.Context, userID, formID string, block Block, parentID *string, orderIndex, depthLevel int, mapping map[string]string) error {
	// Validate
	block.Question = strings.TrimSpace(block.Question)
	if block.Question == "" || block.Type == "" {
		return ErrInvalidInput
	}

	// Find or create question
	questionID, err := s.repo.FindQuestionByText(ctx, block.Type, block.Question)
	if err != nil {
		// Question doesn't exist, create it
		questionID, err = s.repo.CreateQuestion(ctx, userID, block.Type, block.Question)
		if err != nil {
			return err
		}
	}

	// Determine if terminal
	isTerminal := len(block.Children) == 0

	// Create flow connection
	connection, err := s.repo.Create(ctx, formID, questionID, parentID, orderIndex, depthLevel, isTerminal)
	if err != nil {
		return err
	}

	// Store mapping: frontend block ID -> database UUID
	if block.ID != "" {
		mapping[block.ID] = connection.ID
	}

	// Process children recursively
	for i, child := range block.Children {
		if err := s.processBlock(ctx, userID, formID, child, &connection.ID, i, depthLevel+1, mapping); err != nil {
			return err
		}
	}

	return nil
}

func (s *FlowService) GetFlow(ctx context.Context, userID, formID string) ([]FlowConnection, error) {
	// Verify user owns the form
	if err := s.repo.VerifyFormOwnership(ctx, formID, userID); err != nil {
		return nil, err
	}

	return s.repo.GetByFormID(ctx, formID)
}

func (s *FlowService) GetFlowTree(ctx context.Context, userID, formID string) (map[string]interface{}, error) {
	// Verify user owns the form
	if err := s.repo.VerifyFormOwnership(ctx, formID, userID); err != nil {
		return nil, err
	}

	items, err := s.repo.GetFlowWithQuestions(ctx, formID)
	if err != nil {
		return nil, err
	}

	// Build tree structure
	blocks := s.buildTree(items, nil)

	return map[string]interface{}{
		"blocks": blocks,
	}, nil
}

func (s *FlowService) buildTree(items []map[string]interface{}, parentID *string) []map[string]interface{} {
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
