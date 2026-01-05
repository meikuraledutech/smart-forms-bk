package responses

import (
	"context"
	"strings"

	"smart-forms/internal/responses/buffer"

	"github.com/google/uuid"
)

type ResponsesService struct {
	repo   *ResponsesRepository
	buffer *buffer.ResponseBuffer
}

func NewResponsesService(repo *ResponsesRepository, buf *buffer.ResponseBuffer) *ResponsesService {
	return &ResponsesService{
		repo:   repo,
		buffer: buf,
	}
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

	// Generate response ID immediately
	responseID := uuid.New().String()

	// Prepare answers for buffering
	answers := make([]buffer.AnswerData, len(req.Responses))
	for i, answer := range req.Responses {
		answers[i] = buffer.AnswerData{
			ResponseID:       responseID,
			FlowConnectionID: answer.FlowConnectionID,
			AnswerText:       answer.AnswerText,
			AnswerValue:      answer.AnswerValue,
			TimeSpent:        answer.TimeSpent,
		}
	}

	// Enqueue for batch processing
	responseData := buffer.ResponseData{
		ResponseID:     responseID,
		FormID:         formID,
		TotalTimeSpent: req.Metadata.TotalTimeSpent,
		FlowPath:       req.Metadata.FlowPath,
		Metadata:       nil,
		Answers:        answers,
	}

	err = s.buffer.Enqueue(responseData)
	if err != nil {
		return "", err
	}

	// Return immediately to user (data will be inserted in batch)
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
	// Get response
	response, err := s.repo.GetResponseByID(ctx, responseID)
	if err != nil {
		return nil, nil, ErrFormNotFound
	}

	// Get all answers for the response
	answers, err := s.repo.GetAnswersByResponseID(ctx, responseID)
	if err != nil {
		return nil, nil, err
	}

	return response, answers, nil
}
