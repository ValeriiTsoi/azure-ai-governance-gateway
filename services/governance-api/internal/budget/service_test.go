package budget

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct {
	policy      Policy
	policyFound bool
	spend       SpendSnapshot

	recorded *Decision
}

func (r *fakeRepository) FindActivePolicy(
	context.Context,
	string,
) (Policy, bool, error) {
	return r.policy, r.policyFound, nil
}

func (r *fakeRepository) CurrentSpend(
	context.Context,
	string,
	time.Time,
	time.Time,
) (SpendSnapshot, error) {
	return r.spend, nil
}

func (r *fakeRepository) RecordDecision(
	_ context.Context,
	_ string,
	decision Decision,
) error {
	value := decision
	r.recorded = &value

	return nil
}

func fixedTime() time.Time {
	return time.Date(
		2026,
		time.September,
		5,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func testPolicy() Policy {
	return Policy{
		ID:                     1,
		PolicyName:             PolicyName,
		CostCenter:             "AI-DEMO",
		Currency:               "USD",
		MonthlyLimitUSD:        10,
		ReviewThresholdPercent: 80,
	}
}

func TestEvaluateAllowsWithinBudget(
	t *testing.T,
) {
	repository := &fakeRepository{
		policy:      testPolicy(),
		policyFound: true,
		spend: SpendSnapshot{
			SpentUSD: 2,
		},
	}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID:  "req_test",
			CostCenter: "AI-DEMO",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "allow" {
		t.Fatalf(
			"expected allow, got %q",
			decision.Decision,
		)
	}

	if decision.UtilizationPercent == nil ||
		*decision.UtilizationPercent != 20 {
		t.Fatalf(
			"unexpected utilization: %#v",
			decision.UtilizationPercent,
		)
	}
}

func TestEvaluateReviewsAtThreshold(
	t *testing.T,
) {
	repository := &fakeRepository{
		policy:      testPolicy(),
		policyFound: true,
		spend: SpendSnapshot{
			SpentUSD: 8,
		},
	}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID:  "req_test",
			CostCenter: "AI-DEMO",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "review" {
		t.Fatalf(
			"expected review, got %q",
			decision.Decision,
		)
	}
}

func TestEvaluateDeniesExceededBudget(
	t *testing.T,
) {
	repository := &fakeRepository{
		policy:      testPolicy(),
		policyFound: true,
		spend: SpendSnapshot{
			SpentUSD: 10,
		},
	}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID:  "req_test",
			CostCenter: "AI-DEMO",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "deny" {
		t.Fatalf(
			"expected deny, got %q",
			decision.Decision,
		)
	}
}

func TestEvaluateDeniesZeroLimit(
	t *testing.T,
) {
	policy := testPolicy()
	policy.MonthlyLimitUSD = 0

	repository := &fakeRepository{
		policy:      policy,
		policyFound: true,
	}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID:  "req_test",
			CostCenter: "AI-DEMO",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "deny" {
		t.Fatalf(
			"expected deny, got %q",
			decision.Decision,
		)
	}
}

func TestEvaluateReviewsUnknownCost(
	t *testing.T,
) {
	repository := &fakeRepository{
		policy:      testPolicy(),
		policyFound: true,
		spend: SpendSnapshot{
			SpentUSD:           1,
			UnknownCostRecords: 1,
		},
	}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID:  "req_test",
			CostCenter: "AI-DEMO",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "review" {
		t.Fatalf(
			"expected review, got %q",
			decision.Decision,
		)
	}
}

func TestEvaluateReviewsMissingPolicy(
	t *testing.T,
) {
	repository := &fakeRepository{}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID:  "req_test",
			CostCenter: "UNMANAGED",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "review" {
		t.Fatalf(
			"expected review, got %q",
			decision.Decision,
		)
	}
}

func TestEvaluateReviewsMissingCostCenter(
	t *testing.T,
) {
	repository := &fakeRepository{}

	service := newServiceWithClock(
		repository,
		fixedTime,
	)

	decision, err := service.Evaluate(
		context.Background(),
		EvaluateInput{
			RequestID: "req_test",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decision.Decision != "review" {
		t.Fatalf(
			"expected review, got %q",
			decision.Decision,
		)
	}
}
