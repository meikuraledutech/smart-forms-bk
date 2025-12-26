package flows

import (
	"context"
	"strings"
)

type FlowService struct {
	repo *FlowRepository
}

func NewFlowService(repo *FlowRepository) *FlowService {
	return &FlowService{repo: repo}
}

func (s *FlowService) UpdateFlow(ctx context.Context, userID, formID string, req FlowRequest) (map[string]string, error) {
	if len(req.Blocks) == 0 {
		return nil, ErrInvalidInput
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

func (s *FlowService) GetFlow(ctx context.Context, formID string) ([]FlowConnection, error) {
	return s.repo.GetByFormID(ctx, formID)
}
