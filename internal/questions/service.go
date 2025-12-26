package questions

import (
	"context"
	"strings"
)

const (
	TypeQuestion = "question"
	TypeOption   = "option"
)

type QuestionService struct {
	repo *QuestionRepository
}

func NewQuestionService(repo *QuestionRepository) *QuestionService {
	return &QuestionService{repo: repo}
}

func (s *QuestionService) Create(ctx context.Context, userID, qType, text, inputType string, validationRules, metadata map[string]any) (*Question, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrInvalidInput
	}

	if !isValidType(qType) {
		return nil, ErrInvalidType
	}

	if validationRules == nil {
		validationRules = make(map[string]any)
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}

	return s.repo.Create(ctx, userID, qType, text, inputType, validationRules, metadata)
}

func (s *QuestionService) GetByID(ctx context.Context, id string) (*Question, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *QuestionService) List(ctx context.Context, qType string, limit, offset int) ([]Question, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, qType, limit, offset)
}

func (s *QuestionService) Update(ctx context.Context, id, qType, text, inputType string, validationRules, metadata map[string]any) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrInvalidInput
	}

	if !isValidType(qType) {
		return ErrInvalidType
	}

	if validationRules == nil {
		validationRules = make(map[string]any)
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}

	return s.repo.Update(ctx, id, qType, text, inputType, validationRules, metadata)
}

func (s *QuestionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func isValidType(qType string) bool {
	return qType == TypeQuestion || qType == TypeOption
}
