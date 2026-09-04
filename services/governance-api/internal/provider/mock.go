package provider

import (
	"context"
	"unicode/utf8"
)

type Mock struct{}

func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) Name() string {
	return "mock"
}

func (m *Mock) Invoke(
	ctx context.Context,
	request InvokeRequest,
) (InvokeResponse, error) {
	if err := ctx.Err(); err != nil {
		return InvokeResponse{}, err
	}

	const content = "Mock provider response: governed AI invocation completed."

	inputTokens := estimateTokens(request.Prompt)
	outputTokens := estimateTokens(content)

	return InvokeResponse{
		Content: content,
		Model:   request.Model,
		Usage: Usage{
			InputTokens:      inputTokens,
			OutputTokens:     outputTokens,
			EstimatedCostUSD: 0,
		},
	}, nil
}

func estimateTokens(value string) int64 {
	characters := utf8.RuneCountInString(value)

	if characters == 0 {
		return 0
	}

	return int64((characters + 3) / 4)
}
