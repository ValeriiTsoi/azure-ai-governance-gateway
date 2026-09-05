package finops

import (
	"math"
	"testing"
)

func TestCalculatorEstimatesCachedAndNonCachedInput(
	t *testing.T,
) {
	catalog, err := NewStaticCatalog(
		[]Rate{
			{
				Provider:                 "test-provider",
				Model:                    "test-model",
				InputPerMillionUSD:       2,
				CachedInputPerMillionUSD: 0.2,
				OutputPerMillionUSD:      10,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	calculator, err := NewCalculator(catalog)
	if err != nil {
		t.Fatal(err)
	}

	estimate, err := calculator.Estimate(
		"test-provider",
		"test-model",
		Usage{
			InputTokens:       250_000,
			CachedInputTokens: 50_000,
			OutputTokens:      100_000,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !estimate.Known {
		t.Fatal("expected known cost estimate")
	}

	if estimate.NonCachedInputTokens != 200_000 {
		t.Fatalf(
			"unexpected non-cached input tokens: %d",
			estimate.NonCachedInputTokens,
		)
	}

	const expectedInput = 0.4
	const expectedCachedInput = 0.01
	const expectedOutput = 1.0
	const expectedTotal = 1.41
	const tolerance = 0.000000001

	if math.Abs(
		estimate.InputCostUSD-expectedInput,
	) > tolerance {
		t.Fatalf(
			"unexpected input cost: %.9f",
			estimate.InputCostUSD,
		)
	}

	if math.Abs(
		estimate.CachedInputCostUSD-expectedCachedInput,
	) > tolerance {
		t.Fatalf(
			"unexpected cached input cost: %.9f",
			estimate.CachedInputCostUSD,
		)
	}

	if math.Abs(
		estimate.OutputCostUSD-expectedOutput,
	) > tolerance {
		t.Fatalf(
			"unexpected output cost: %.9f",
			estimate.OutputCostUSD,
		)
	}

	if math.Abs(
		estimate.TotalCostUSD-expectedTotal,
	) > tolerance {
		t.Fatalf(
			"unexpected total cost: %.9f",
			estimate.TotalCostUSD,
		)
	}
}

func TestCalculatorReturnsUnknownForMissingPrice(
	t *testing.T,
) {
	catalog, err := NewStaticCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}

	calculator, err := NewCalculator(catalog)
	if err != nil {
		t.Fatal(err)
	}

	estimate, err := calculator.Estimate(
		"unknown-provider",
		"unknown-model",
		Usage{
			InputTokens:       100,
			CachedInputTokens: 20,
			OutputTokens:      50,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if estimate.Known {
		t.Fatal("expected unknown cost estimate")
	}
}

func TestCalculatorRejectsNegativeTokenUsage(
	t *testing.T,
) {
	catalog, err := NewStaticCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}

	calculator, err := NewCalculator(catalog)
	if err != nil {
		t.Fatal(err)
	}

	_, err = calculator.Estimate(
		"test-provider",
		"test-model",
		Usage{
			InputTokens: -1,
		},
	)

	if err == nil {
		t.Fatal("expected negative token usage error")
	}
}

func TestCalculatorRejectsCachedTokensAboveInput(
	t *testing.T,
) {
	catalog, err := NewStaticCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}

	calculator, err := NewCalculator(catalog)
	if err != nil {
		t.Fatal(err)
	}

	_, err = calculator.Estimate(
		"test-provider",
		"test-model",
		Usage{
			InputTokens:       10,
			CachedInputTokens: 11,
		},
	)

	if err == nil {
		t.Fatal(
			"expected cached token validation error",
		)
	}
}
