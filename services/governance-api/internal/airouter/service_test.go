package airouter

import (
	"context"
	"errors"
	"testing"

	"governance-api/internal/budget"
	"governance-api/internal/finops"
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

type fakeBudgetService struct {
	decision string
	calls    int
}

func (s *fakeBudgetService) Evaluate(
	_ context.Context,
	input budget.EvaluateInput,
) (budget.Decision, error) {
	s.calls++

	decision := s.decision
	if decision == "" {
		decision = "allow"
	}

	return budget.Decision{
		PolicyName: budget.PolicyName,
		Decision:   decision,
		Reason:     "test budget decision",
		CostCenter: input.CostCenter,
		Currency:   "USD",
	}, nil
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
			InputTokens:       10,
			CachedInputTokens: 4,
			OutputTokens:      5,
		},
	}, nil
}

func newTestCostCalculator() *finops.Calculator {
	catalog, err := finops.NewStaticCatalog(
		[]finops.Rate{
			{
				Provider:                 "mock",
				Model:                    "mock-fast-general",
				InputPerMillionUSD:       2,
				CachedInputPerMillionUSD: 0.2,
				OutputPerMillionUSD:      10,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	calculator, err := finops.NewCalculator(catalog)
	if err != nil {
		panic(err)
	}

	return calculator
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
		&fakeBudgetService{decision: "allow"},
		repository,
		newTestCostCalculator(),
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

func TestInvokeAllowCalculatesCostWithFinOps(
	t *testing.T,
) {
	service, _, _ := newTestService("allow")

	result, err := service.Invoke(
		context.Background(),
		InvokeInput{
			CallerSubject:      "finops-test@example.com",
			DataClassification: "internal",
			RequestedModel:     "fast-general",
			Prompt:             "hello model",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Usage == nil {
		t.Fatal("expected usage")
	}

	if result.Usage.EstimatedCostUSD == nil {
		t.Fatal("expected known FinOps cost")
	}

	// fakeProvider:
	// input  = 10 tokens
	// output = 5 tokens
	//
	// pricing:
	// input  = $2 / 1M
	// output = $10 / 1M
	//
	// 10/1M*2 + 5/1M*10 = $0.00007
	// 6 non-cached input * $2 / 1M
	// + 4 cached input * $0.2 / 1M
	// + 5 output * $10 / 1M
	// = $0.0000628
	const expected = 0.0000628
	const tolerance = 0.000000000001

	actual := *result.Usage.EstimatedCostUSD

	difference := actual - expected
	if difference < 0 {
		difference = -difference
	}

	if difference > tolerance {
		t.Fatalf(
			"unexpected estimated cost: %.12f",
			actual,
		)
	}
}

func TestInvokeAllowReturnsPricingSnapshot(
	t *testing.T,
) {
	service, _, _ := newTestService("allow")

	result, err := service.Invoke(
		context.Background(),
		InvokeInput{
			CallerSubject:      "pricing-audit@example.com",
			DataClassification: "internal",
			RequestedModel:     "fast-general",
			Prompt:             "pricing audit test",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Usage == nil {
		t.Fatal("expected usage")
	}

	if result.Usage.Pricing == nil {
		t.Fatal("expected pricing snapshot")
	}

	if result.Usage.Pricing.InputPerMillionUSD != 2 {
		t.Fatalf(
			"unexpected input price: %f",
			result.Usage.Pricing.InputPerMillionUSD,
		)
	}

	if result.Usage.Pricing.CachedInputPerMillionUSD != 0.2 {
		t.Fatalf(
			"unexpected cached input price: %f",
			result.Usage.Pricing.CachedInputPerMillionUSD,
		)
	}

	if result.Usage.Pricing.OutputPerMillionUSD != 10 {
		t.Fatalf(
			"unexpected output price: %f",
			result.Usage.Pricing.OutputPerMillionUSD,
		)
	}
}

func newBudgetGuardrailTestService(
	governanceDecision string,
	budgetDecision string,
) (
	*Service,
	*fakeBudgetService,
	*fakeProvider,
	*fakeRepository,
) {
	modelProvider := &fakeProvider{}
	repository := &fakeRepository{}
	budgetService := &fakeBudgetService{
		decision: budgetDecision,
	}

	service := NewService(
		&fakeGovernanceService{
			decision: governanceDecision,
		},
		budgetService,
		repository,
		newTestCostCalculator(),
		map[string]provider.Provider{
			"mock": modelProvider,
		},
		map[string]Route{
			"fast-general": {
				RoutedModel: "mock-fast-general",
				Provider:    "mock",
				Reason:      "budget guardrail test route",
			},
		},
	)

	return service, budgetService, modelProvider, repository
}

func budgetGuardrailInvokeInput() InvokeInput {
	return InvokeInput{
		CallerSubject:      "budget-test@example.com",
		CostCenter:         "BUDGET-TEST",
		UseCase:            "budget-guardrail-test",
		DataClassification: "internal",
		RequestedModel:     "fast-general",
		Prompt:             "synthetic budget guardrail test",
	}
}

func TestBudgetGuardrailGovernanceDenyStopsBeforeBudget(
	t *testing.T,
) {
	service, budgetService, modelProvider, _ :=
		newBudgetGuardrailTestService(
			"deny",
			"allow",
		)

	result, err := service.Invoke(
		context.Background(),
		budgetGuardrailInvokeInput(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if budgetService.calls != 0 {
		t.Fatalf(
			"expected zero budget evaluations, got %d",
			budgetService.calls,
		)
	}

	if modelProvider.calls != 0 {
		t.Fatalf(
			"expected zero provider calls, got %d",
			modelProvider.calls,
		)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be marked as called")
	}

	if result.Budget != nil {
		t.Fatal(
			"budget must not be evaluated after governance deny",
		)
	}

	if result.Route != nil ||
		result.Response != nil ||
		result.Usage != nil {
		t.Fatal(
			"deny must not create route, response or usage",
		)
	}
}

func TestBudgetGuardrailAllowContinuesToProvider(
	t *testing.T,
) {
	service, budgetService, modelProvider, _ :=
		newBudgetGuardrailTestService(
			"allow",
			"allow",
		)

	result, err := service.Invoke(
		context.Background(),
		budgetGuardrailInvokeInput(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if budgetService.calls != 1 {
		t.Fatalf(
			"expected one budget evaluation, got %d",
			budgetService.calls,
		)
	}

	if modelProvider.calls != 1 {
		t.Fatalf(
			"expected one provider call, got %d",
			modelProvider.calls,
		)
	}

	if result.Budget == nil ||
		result.Budget.Decision != "allow" {
		t.Fatalf(
			"expected budget allow, got %#v",
			result.Budget,
		)
	}

	if !result.ProviderCalled {
		t.Fatal("expected provider_called=true")
	}

	if result.Route == nil {
		t.Fatal("expected model route")
	}

	if result.Response == nil {
		t.Fatal("expected model response")
	}

	if result.Usage == nil {
		t.Fatal("expected usage")
	}
}

func TestBudgetGuardrailReviewStopsBeforeRoutingAndProvider(
	t *testing.T,
) {
	service, budgetService, modelProvider, _ :=
		newBudgetGuardrailTestService(
			"allow",
			"review",
		)

	result, err := service.Invoke(
		context.Background(),
		budgetGuardrailInvokeInput(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if budgetService.calls != 1 {
		t.Fatalf(
			"expected one budget evaluation, got %d",
			budgetService.calls,
		)
	}

	if modelProvider.calls != 0 {
		t.Fatalf(
			"expected zero provider calls, got %d",
			modelProvider.calls,
		)
	}

	if result.Budget == nil ||
		result.Budget.Decision != "review" {
		t.Fatalf(
			"expected budget review, got %#v",
			result.Budget,
		)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be called")
	}

	if result.Route != nil ||
		result.Response != nil ||
		result.Usage != nil {
		t.Fatal(
			"budget review must stop before route/provider/usage",
		)
	}
}

func TestBudgetGuardrailDenyStopsBeforeRoutingAndProvider(
	t *testing.T,
) {
	service, budgetService, modelProvider, _ :=
		newBudgetGuardrailTestService(
			"allow",
			"deny",
		)

	result, err := service.Invoke(
		context.Background(),
		budgetGuardrailInvokeInput(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if budgetService.calls != 1 {
		t.Fatalf(
			"expected one budget evaluation, got %d",
			budgetService.calls,
		)
	}

	if modelProvider.calls != 0 {
		t.Fatalf(
			"expected zero provider calls, got %d",
			modelProvider.calls,
		)
	}

	if result.Budget == nil ||
		result.Budget.Decision != "deny" {
		t.Fatalf(
			"expected budget deny, got %#v",
			result.Budget,
		)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be called")
	}

	if result.Route != nil ||
		result.Response != nil ||
		result.Usage != nil {
		t.Fatal(
			"budget deny must stop before route/provider/usage",
		)
	}
}
