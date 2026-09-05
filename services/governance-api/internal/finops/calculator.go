package finops

import (
	"errors"
	"fmt"
)

const tokensPerMillion = 1_000_000.0

type Usage struct {
	InputTokens  int64
	OutputTokens int64
}

type CostEstimate struct {
	Known         bool
	InputCostUSD  float64
	OutputCostUSD float64
	TotalCostUSD  float64
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

	if usage.OutputTokens < 0 {
		return CostEstimate{}, fmt.Errorf(
			"output tokens must be non-negative: %d",
			usage.OutputTokens,
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

	inputCost :=
		float64(usage.InputTokens) /
			tokensPerMillion *
			rate.InputPerMillionUSD

	outputCost :=
		float64(usage.OutputTokens) /
			tokensPerMillion *
			rate.OutputPerMillionUSD

	return CostEstimate{
		Known:         true,
		InputCostUSD:  inputCost,
		OutputCostUSD: outputCost,
		TotalCostUSD:  inputCost + outputCost,
	}, nil
}
