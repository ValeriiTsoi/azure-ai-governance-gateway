package finops

import (
	"errors"
	"fmt"
	"strings"
)

type Rate struct {
	Provider            string
	Model               string
	InputPerMillionUSD  float64
	OutputPerMillionUSD float64
}

type Catalog interface {
	Lookup(
		provider string,
		model string,
	) (Rate, bool)
}

type StaticCatalog struct {
	rates map[string]Rate
}

func NewStaticCatalog(
	rates []Rate,
) (*StaticCatalog, error) {
	result := &StaticCatalog{
		rates: make(map[string]Rate, len(rates)),
	}

	for _, rate := range rates {
		provider := strings.TrimSpace(rate.Provider)
		model := strings.TrimSpace(rate.Model)

		if provider == "" {
			return nil, errors.New(
				"pricing provider is required",
			)
		}

		if model == "" {
			return nil, errors.New(
				"pricing model is required",
			)
		}

		if rate.InputPerMillionUSD < 0 {
			return nil, fmt.Errorf(
				"negative input price for %s/%s",
				provider,
				model,
			)
		}

		if rate.OutputPerMillionUSD < 0 {
			return nil, fmt.Errorf(
				"negative output price for %s/%s",
				provider,
				model,
			)
		}

		key := pricingKey(provider, model)

		if _, exists := result.rates[key]; exists {
			return nil, fmt.Errorf(
				"duplicate pricing rate for %s/%s",
				provider,
				model,
			)
		}

		rate.Provider = provider
		rate.Model = model

		result.rates[key] = rate
	}

	return result, nil
}

func (c *StaticCatalog) Lookup(
	provider string,
	model string,
) (Rate, bool) {
	if c == nil {
		return Rate{}, false
	}

	key := pricingKey(provider, model)

	rate, ok := c.rates[key]

	return rate, ok
}

func pricingKey(
	provider string,
	model string,
) string {
	return strings.ToLower(
		strings.TrimSpace(provider),
	) + "\x00" + strings.ToLower(
		strings.TrimSpace(model),
	)
}
