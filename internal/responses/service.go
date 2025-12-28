package responses

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type ResponsesService struct {
	repo *ResponsesRepository
}

func NewResponsesService(repo *ResponsesRepository) *ResponsesService {
	return &ResponsesService{repo: repo}
}

// SubmitResponse handles form response submission
func (s *ResponsesService) SubmitResponse(ctx context.Context, slug string, req SubmitRequest) (string, error) {
	// Validate input
	if len(req.Responses) == 0 {
		return "", ErrInvalidInput
	}

	if req.Metadata.TotalTimeSpent < 0 {
		return "", ErrInvalidInput
	}

	if len(req.Metadata.FlowPath) == 0 {
		return "", ErrInvalidInput
	}

	// Get form by slug
	formID, acceptingResponses, err := s.repo.GetFormBySlug(ctx, slug)
	if err != nil {
		return "", ErrFormNotFound
	}

	// Check if form is accepting responses
	if !acceptingResponses {
		return "", ErrFormNotAccepting
	}

	// Verify all flow_connection_ids exist
	for _, answer := range req.Responses {
		if strings.TrimSpace(answer.FlowConnectionID) == "" {
			return "", ErrInvalidInput
		}

		if strings.TrimSpace(answer.AnswerText) == "" {
			return "", ErrInvalidInput
		}

		// Validate UUID format
		if _, err := uuid.Parse(answer.FlowConnectionID); err != nil {
			return "", ErrInvalidFlowConnection
		}

		err := s.repo.VerifyFlowConnection(ctx, formID, answer.FlowConnectionID)
		if err != nil {
			return "", err
		}
	}

	// Create response record
	responseID, err := s.repo.CreateResponse(ctx, formID, req.Metadata.TotalTimeSpent, req.Metadata.FlowPath, nil)
	if err != nil {
		return "", err
	}

	// Create answer records
	for _, answer := range req.Responses {
		err := s.repo.CreateAnswer(ctx, responseID, answer.FlowConnectionID, answer.AnswerText, answer.AnswerValue, answer.TimeSpent)
		if err != nil {
			return "", err
		}
	}

	return responseID, nil
}

// GetResponses retrieves responses for a form (owner only)
func (s *ResponsesService) GetResponses(ctx context.Context, formID string, limit, offset int) ([]FormResponse, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetResponsesByFormID(ctx, formID, limit, offset)
}

// GetResponseDetails retrieves a single response with all answers
func (s *ResponsesService) GetResponseDetails(ctx context.Context, responseID string) (*FormResponse, []ResponseAnswer, error) {
	// This would need a repo method to get single response
	// For now, returning error as placeholder
	return nil, nil, ErrInvalidInput
}
