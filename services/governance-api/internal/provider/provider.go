package provider

import "context"

type InvokeRequest struct {
	Model  string
	Prompt string
}

type Usage struct {
	InputTokens      int64
	OutputTokens     int64
	EstimatedCostUSD float64
}

type InvokeResponse struct {
	Content string
	Model   string
	Usage   Usage
}

type Provider interface {
	Name() string

	Invoke(
		context.Context,
		InvokeRequest,
	) (InvokeResponse, error)
}
