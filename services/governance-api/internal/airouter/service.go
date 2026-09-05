package airouter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"governance-api/internal/finops"
	"governance-api/internal/governance"
	"governance-api/internal/provider"
)

var (
	ErrUnsupportedModel           = errors.New("unsupported requested model")
	ErrProviderNotConfigured      = errors.New("model provider is not configured")
	ErrProviderInvocation         = errors.New("model provider invocation failed")
	ErrCostEstimatorNotConfigured = errors.New("FinOps cost estimator is not configured")
	ErrCostEstimation             = errors.New("AI invocation cost estimation failed")
)

type GovernanceService interface {
	CreateRequest(
		context.Context,
		governance.CreateRequestInput,
	) (governance.Request, error)
}

type Repository interface {
	RecordInvocation(
		context.Context,
		string,
		Route,
		Usage,
	) error
}

type CostEstimator interface {
	Estimate(
		provider string,
		model string,
		usage finops.Usage,
	) (finops.CostEstimate, error)
}

type Service struct {
	governance    GovernanceService
	repository    Repository
	costEstimator CostEstimator
	providers     map[string]provider.Provider
	routes        map[string]Route
}

func NewService(
	governanceService GovernanceService,
	repository Repository,
	costEstimator CostEstimator,
	providers map[string]provider.Provider,
	routes map[string]Route,
) *Service {
	providerRegistry := make(
		map[string]provider.Provider,
		len(providers),
	)

	for name, modelProvider := range providers {
		providerRegistry[strings.TrimSpace(name)] = modelProvider
	}

	routeRegistry := make(
		map[string]Route,
		len(routes),
	)

	for name, route := range routes {
		routeRegistry[strings.TrimSpace(name)] = route
	}

	return &Service{
		governance:    governanceService,
		repository:    repository,
		costEstimator: costEstimator,
		providers:     providerRegistry,
		routes:        routeRegistry,
	}
}

func (s *Service) Invoke(
	ctx context.Context,
	input InvokeInput,
) (Result, error) {
	governanceRequest, err := s.governance.CreateRequest(
		ctx,
		governance.CreateRequestInput{
			CallerSubject:      input.CallerSubject,
			CostCenter:         input.CostCenter,
			UseCase:            input.UseCase,
			DataClassification: input.DataClassification,
			RequestedModel:     input.RequestedModel,
			Prompt:             input.Prompt,
			Metadata:           input.Metadata,
		},
	)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Governance:     governanceRequest,
		ProviderCalled: false,
	}

	switch governanceRequest.Policy.Decision {
	case "deny", "review":
		return result, nil

	case "allow":
		// Continue to routing and provider invocation.

	default:
		return Result{}, fmt.Errorf(
			"unsupported governance decision %q",
			governanceRequest.Policy.Decision,
		)
	}

	requestedModel := strings.TrimSpace(
		governanceRequest.RequestedModel,
	)

	route, ok := s.routes[requestedModel]
	if !ok {
		return result, fmt.Errorf(
			"%w: %s",
			ErrUnsupportedModel,
			requestedModel,
		)
	}

	route.RequestedModel = requestedModel

	modelProvider, ok := s.providers[route.Provider]
	if !ok {
		return result, fmt.Errorf(
			"%w: %s",
			ErrProviderNotConfigured,
			route.Provider,
		)
	}

	providerResponse, err := modelProvider.Invoke(
		ctx,
		provider.InvokeRequest{
			Model:  route.RoutedModel,
			Prompt: input.Prompt,
		},
	)
	if err != nil {
		return result, fmt.Errorf(
			"%w: %v",
			ErrProviderInvocation,
			err,
		)
	}

	result.ProviderCalled = true

	model := providerResponse.Model
	if strings.TrimSpace(model) == "" {
		model = route.RoutedModel
	}

	if s.costEstimator == nil {
		return result, ErrCostEstimatorNotConfigured
	}

	costEstimate, err := s.costEstimator.Estimate(
		route.Provider,
		model,
		finops.Usage{
			InputTokens: providerResponse.Usage.InputTokens,
			CachedInputTokens: providerResponse.
				Usage.CachedInputTokens,
			OutputTokens: providerResponse.Usage.OutputTokens,
		},
	)
	if err != nil {
		return result, fmt.Errorf(
			"%w: %v",
			ErrCostEstimation,
			err,
		)
	}

	var estimatedCostUSD *float64
	if costEstimate.Known {
		value := costEstimate.TotalCostUSD
		estimatedCostUSD = &value
	}

	usage := Usage{
		Provider:    route.Provider,
		Model:       model,
		InputTokens: providerResponse.Usage.InputTokens,
		CachedInputTokens: providerResponse.
			Usage.CachedInputTokens,
		OutputTokens:     providerResponse.Usage.OutputTokens,
		EstimatedCostUSD: estimatedCostUSD,
	}

	if err := s.repository.RecordInvocation(
		ctx,
		governanceRequest.RequestID,
		route,
		usage,
	); err != nil {
		return result, fmt.Errorf(
			"record AI invocation: %w",
			err,
		)
	}

	result.Route = &route
	result.Response = &ModelResponse{
		Provider: route.Provider,
		Model:    model,
		Content:  providerResponse.Content,
	}
	result.Usage = &usage

	return result, nil
}
