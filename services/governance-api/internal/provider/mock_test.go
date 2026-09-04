package provider

import (
	"context"
	"testing"
)

func TestMockInvokeReturnsDeterministicUsage(
	t *testing.T,
) {
	modelProvider := NewMock()

	result, err := modelProvider.Invoke(
		context.Background(),
		InvokeRequest{
			Model:  "mock-fast-general",
			Prompt: "hello world",
		},
	)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if result.Model != "mock-fast-general" {
		t.Fatalf(
			"unexpected model %q",
			result.Model,
		)
	}

	if result.Content == "" {
		t.Fatal("expected non-empty mock content")
	}

	if result.Usage.InputTokens <= 0 {
		t.Fatalf(
			"expected positive input token estimate, got %d",
			result.Usage.InputTokens,
		)
	}

	if result.Usage.OutputTokens <= 0 {
		t.Fatalf(
			"expected positive output token estimate, got %d",
			result.Usage.OutputTokens,
		)
	}

	if result.Usage.EstimatedCostUSD != 0 {
		t.Fatalf(
			"expected zero mock cost, got %f",
			result.Usage.EstimatedCostUSD,
		)
	}
}
