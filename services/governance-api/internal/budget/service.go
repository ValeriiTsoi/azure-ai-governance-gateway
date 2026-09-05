package budget

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New(
	"invalid budget evaluation input",
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func newServiceWithClock(
	repository Repository,
	now func() time.Time,
) *Service {
	return &Service{
		repository: repository,
		now:        now,
	}
}

func (s *Service) Evaluate(
	ctx context.Context,
	input EvaluateInput,
) (Decision, error) {
	requestID := strings.TrimSpace(input.RequestID)
	costCenter := strings.TrimSpace(input.CostCenter)

	if requestID == "" {
		return Decision{}, fmt.Errorf(
			"%w: request ID is required",
			ErrInvalidInput,
		)
	}

	now := s.now().UTC()

	periodStart := time.Date(
		now.Year(),
		now.Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := periodStart.AddDate(0, 1, 0)

	decision := Decision{
		PolicyName: PolicyName,
		Decision:   "review",
		CostCenter: costCenter,
		Currency:   "USD",

		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,

		SpentUSD:           0,
		UnknownCostRecords: 0,
		EvaluatedAt:        now,
	}

	if costCenter == "" {
		decision.Reason =
			"cost center is required for budget enforcement"

		return s.record(ctx, requestID, decision)
	}

	policy, found, err :=
		s.repository.FindActivePolicy(
			ctx,
			costCenter,
		)
	if err != nil {
		return Decision{}, err
	}

	if !found {
		decision.Reason =
			"no active budget policy is configured for the cost center"

		return s.record(ctx, requestID, decision)
	}

	policyID := policy.ID
	monthlyLimit := policy.MonthlyLimitUSD
	reviewThreshold := policy.ReviewThresholdPercent

	decision.PolicyID = &policyID
	decision.PolicyName = policy.PolicyName
	decision.Currency = policy.Currency
	decision.MonthlyLimitUSD = &monthlyLimit
	decision.ReviewThresholdPercent = &reviewThreshold

	snapshot, err := s.repository.CurrentSpend(
		ctx,
		costCenter,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return Decision{}, err
	}

	decision.SpentUSD = snapshot.SpentUSD
	decision.UnknownCostRecords =
		snapshot.UnknownCostRecords

	if snapshot.UnknownCostRecords > 0 {
		decision.Decision = "review"
		decision.Reason =
			"current budget period contains usage with unknown cost"

		return s.record(ctx, requestID, decision)
	}

	if monthlyLimit == 0 {
		zero := float64(100)

		decision.UtilizationPercent = &zero
		decision.Decision = "deny"
		decision.Reason =
			"monthly budget limit is zero"

		return s.record(ctx, requestID, decision)
	}

	utilization :=
		(snapshot.SpentUSD / monthlyLimit) * 100

	decision.UtilizationPercent = &utilization

	switch {
	case snapshot.SpentUSD >= monthlyLimit:
		decision.Decision = "deny"
		decision.Reason =
			"monthly budget limit has been reached or exceeded"

	case utilization >= reviewThreshold:
		decision.Decision = "review"
		decision.Reason =
			"monthly budget review threshold has been reached"

	default:
		decision.Decision = "allow"
		decision.Reason =
			"monthly budget is within the configured threshold"
	}

	return s.record(ctx, requestID, decision)
}

func (s *Service) record(
	ctx context.Context,
	requestID string,
	decision Decision,
) (Decision, error) {
	if err := s.repository.RecordDecision(
		ctx,
		requestID,
		decision,
	); err != nil {
		return Decision{}, err
	}

	return decision, nil
}
