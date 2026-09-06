package finops

import (
	"errors"
	"fmt"
)

const tokensPerMillion = 1_000_000.0

type Usage struct {
	InputTokens       int64
	CachedInputTokens int64
	OutputTokens      int64
}

type CostEstimate struct {
	Known                bool
	NonCachedInputTokens int64
	CachedInputTokens    int64
	InputCostUSD         float64
	CachedInputCostUSD   float64
	OutputCostUSD        float64
	TotalCostUSD         float64
	Rate                 Rate
}

type Calculator struct {
	catalog Catalog
}

func NewCalculator(
	catalog Catalog,
) (*Calculator, error) {
	if catalog == nil {
		return nil, errors.New(
			"pricing catalog is required",
		)
	}

	return &Calculator{
		catalog: catalog,
	}, nil
}

func (c *Calculator) Estimate(
	provider string,
	model string,
	usage Usage,
) (CostEstimate, error) {
	if usage.InputTokens < 0 {
		return CostEstimate{}, fmt.Errorf(
			"input tokens must be non-negative: %d",
			usage.InputTokens,
		)
	}

	if usage.CachedInputTokens < 0 {
		return CostEstimate{}, fmt.Errorf(
			"cached input tokens must be non-negative: %d",
			usage.CachedInputTokens,
		)
	}

	if usage.OutputTokens < 0 {
		return CostEstimate{}, fmt.Errorf(
			"output tokens must be non-negative: %d",
			usage.OutputTokens,
		)
	}

	if usage.CachedInputTokens > usage.InputTokens {
		return CostEstimate{}, fmt.Errorf(
			"cached input tokens %d exceed input tokens %d",
			usage.CachedInputTokens,
			usage.InputTokens,
		)
	}

	rate, found := c.catalog.Lookup(
		provider,
		model,
	)

	if !found {
		return CostEstimate{
			Known: false,
		}, nil
	}

	nonCachedInputTokens :=
		usage.InputTokens - usage.CachedInputTokens

	inputCost :=
		float64(nonCachedInputTokens) /
			tokensPerMillion *
			rate.InputPerMillionUSD

	cachedInputCost :=
		float64(usage.CachedInputTokens) /
			tokensPerMillion *
			rate.CachedInputPerMillionUSD

	outputCost :=
		float64(usage.OutputTokens) /
			tokensPerMillion *
			rate.OutputPerMillionUSD

	return CostEstimate{
		Known:                true,
		NonCachedInputTokens: nonCachedInputTokens,
		CachedInputTokens:    usage.CachedInputTokens,
		InputCostUSD:         inputCost,
		CachedInputCostUSD:   cachedInputCost,
		OutputCostUSD:        outputCost,
		TotalCostUSD:         inputCost + cachedInputCost + outputCost,
		Rate:                 rate,
	}, nil
}
