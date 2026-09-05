package airouter

import (
	"context"
	"errors"
	"testing"

	"governance-api/internal/governance"
	"governance-api/internal/provider"
)

type fakeGovernanceService struct {
	decision string
}

func (f *fakeGovernanceService) CreateRequest(
	_ context.Context,
	input governance.CreateRequestInput,
) (governance.Request, error) {
	return governance.Request{
		RequestID:          "req_test",
		CallerSubject:      input.CallerSubject,
		DataClassification: input.DataClassification,
		RequestedModel:     input.RequestedModel,
		PromptHash:         "test-hash",
		PromptChars:        len(input.Prompt),
		Metadata:           map[string]any{},
		Policy: governance.PolicyDecision{
			PolicyName: "test-policy",
			Decision:   f.decision,
			Reason:     "test decision",
		},
	}, nil
}

type fakeRepository struct {
	records int
	route   Route
	usage   Usage
}

func (f *fakeRepository) RecordInvocation(
	_ context.Context,
	_ string,
	route Route,
	usage Usage,
) error {
	f.records++
	f.route = route
	f.usage = usage

	return nil
}

type fakeProvider struct {
	calls int
}

func (f *fakeProvider) Name() string {
	return "mock"
}

func (f *fakeProvider) Invoke(
	_ context.Context,
	request provider.InvokeRequest,
) (provider.InvokeResponse, error) {
	f.calls++

	return provider.InvokeResponse{
		Content: "test response",
		Model:   request.Model,
		Usage: provider.Usage{
			InputTokens:      10,
			OutputTokens:     5,
			EstimatedCostUSD: new(float64),
		},
	}, nil
}

func newTestService(
	decision string,
) (*Service, *fakeProvider, *fakeRepository) {
	modelProvider := &fakeProvider{}
	repository := &fakeRepository{}

	service := NewService(
		&fakeGovernanceService{
			decision: decision,
		},
		repository,
		map[string]provider.Provider{
			"mock": modelProvider,
		},
		map[string]Route{
			"fast-general": {
				RoutedModel: "mock-fast-general",
				Provider:    "mock",
				Reason:      "test route",
			},
		},
	)

	return service, modelProvider, repository
}

func TestInvokeAllowCallsProviderAndRecordsUsage(
	t *testing.T,
) {
	service, modelProvider, repository := newTestService(
		"allow",
	)

	result, err := service.Invoke(
		context.Background(),
		InvokeInput{
			CallerSubject:      "test@example.com",
			DataClassification: "internal",
			RequestedModel:     "fast-general",
			Prompt:             "hello model",
		},
	)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if !result.ProviderCalled {
		t.Fatal("expected provider to be called")
	}

	if modelProvider.calls != 1 {
		t.Fatalf(
			"expected provider calls=1, got %d",
			modelProvider.calls,
		)
	}

	if repository.records != 1 {
		t.Fatalf(
			"expected repository records=1, got %d",
			repository.records,
		)
	}

	if result.Route == nil {
		t.Fatal("expected route in result")
	}

	if result.Route.Provider != "mock" {
		t.Fatalf(
			"expected provider mock, got %q",
			result.Route.Provider,
		)
	}

	if result.Route.RoutedModel != "mock-fast-general" {
		t.Fatalf(
			"unexpected routed model %q",
			result.Route.RoutedModel,
		)
	}

	if result.Usage == nil {
		t.Fatal("expected usage in result")
	}

	if result.Usage.InputTokens != 10 {
		t.Fatalf(
			"expected input_tokens=10, got %d",
			result.Usage.InputTokens,
		)
	}
}

func TestInvokeReviewDoesNotCallProvider(
	t *testing.T,
) {
	service, modelProvider, repository := newTestService(
		"review",
	)

	result, err := service.Invoke(
		context.Background(),
		InvokeInput{
			CallerSubject:      "test@example.com",
			DataClassification: "confidential",
			RequestedModel:     "fast-general",
			Prompt:             "confidential prompt",
		},
	)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be called for review")
	}

	if modelProvider.calls != 0 {
		t.Fatalf(
			"expected provider calls=0, got %d",
			modelProvider.calls,
		)
	}

	if repository.records != 0 {
		t.Fatalf(
			"expected repository records=0, got %d",
			repository.records,
		)
	}
}

func TestInvokeDenyDoesNotCallProvider(
	t *testing.T,
) {
	service, modelProvider, repository := newTestService(
		"deny",
	)

	result, err := service.Invoke(
		context.Background(),
		InvokeInput{
			CallerSubject:      "test@example.com",
			DataClassification: "restricted",
			RequestedModel:     "fast-general",
			Prompt:             "restricted prompt",
		},
	)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be called for deny")
	}

	if modelProvider.calls != 0 {
		t.Fatalf(
			"expected provider calls=0, got %d",
			modelProvider.calls,
		)
	}

	if repository.records != 0 {
		t.Fatalf(
			"expected repository records=0, got %d",
			repository.records,
		)
	}
}

func TestInvokeRejectsUnsupportedModel(
	t *testing.T,
) {
	service, modelProvider, repository := newTestService(
		"allow",
	)

	_, err := service.Invoke(
		context.Background(),
		InvokeInput{
			CallerSubject:      "test@example.com",
			DataClassification: "internal",
			RequestedModel:     "unknown-model",
			Prompt:             "hello",
		},
	)

	if !errors.Is(err, ErrUnsupportedModel) {
		t.Fatalf(
			"expected ErrUnsupportedModel, got %v",
			err,
		)
	}

	if modelProvider.calls != 0 {
		t.Fatalf(
			"expected provider calls=0, got %d",
			modelProvider.calls,
		)
	}

	if repository.records != 0 {
		t.Fatalf(
			"expected repository records=0, got %d",
			repository.records,
		)
	}
}
