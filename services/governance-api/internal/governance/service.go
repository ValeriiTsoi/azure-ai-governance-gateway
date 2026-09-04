package governance

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidInput = errors.New("invalid governance request")
	ErrNotFound     = errors.New("governance request not found")
)

type Repository interface {
	Create(
		context.Context,
		Request,
		PolicyDecision,
	) (Request, error)

	GetByRequestID(
		context.Context,
		string,
	) (Request, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateRequest(
	ctx context.Context,
	input CreateRequestInput,
) (Request, error) {
	if err := validateCreateRequest(input); err != nil {
		return Request{}, err
	}

	requestID, err := newRequestID()
	if err != nil {
		return Request{}, fmt.Errorf("generate request id: %w", err)
	}

	hash := sha256.Sum256([]byte(input.Prompt))

	request := Request{
		RequestID:          requestID,
		CallerSubject:      strings.TrimSpace(input.CallerSubject),
		CostCenter:         strings.TrimSpace(input.CostCenter),
		UseCase:            strings.TrimSpace(input.UseCase),
		DataClassification: input.DataClassification,
		RequestedModel:     strings.TrimSpace(input.RequestedModel),
		PromptHash:         hex.EncodeToString(hash[:]),
		PromptChars:        utf8.RuneCountInString(input.Prompt),
		Metadata:           input.Metadata,
	}

	if request.Metadata == nil {
		request.Metadata = map[string]any{}
	}

	decision := EvaluatePolicy(request.DataClassification)

	return s.repository.Create(ctx, request, decision)
}

func (s *Service) GetRequest(
	ctx context.Context,
	requestID string,
) (Request, error) {
	requestID = strings.TrimSpace(requestID)

	if requestID == "" {
		return Request{}, fmt.Errorf(
			"%w: request_id is required",
			ErrInvalidInput,
		)
	}

	return s.repository.GetByRequestID(ctx, requestID)
}

func validateCreateRequest(input CreateRequestInput) error {
	if strings.TrimSpace(input.CallerSubject) == "" {
		return fmt.Errorf(
			"%w: caller_subject is required",
			ErrInvalidInput,
		)
	}

	if strings.TrimSpace(input.RequestedModel) == "" {
		return fmt.Errorf(
			"%w: requested_model is required",
			ErrInvalidInput,
		)
	}

	if input.Prompt == "" {
		return fmt.Errorf(
			"%w: prompt is required",
			ErrInvalidInput,
		)
	}

	switch input.DataClassification {
	case "public", "internal", "confidential", "restricted":
	default:
		return fmt.Errorf(
			"%w: unsupported data_classification",
			ErrInvalidInput,
		)
	}

	return nil
}

func newRequestID() (string, error) {
	random := make([]byte, 16)

	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	return "req_" + hex.EncodeToString(random), nil
}
