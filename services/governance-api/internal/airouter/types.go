package airouter

import "governance-api/internal/governance"

type InvokeInput struct {
	CallerSubject      string         `json:"caller_subject"`
	CostCenter         string         `json:"cost_center,omitempty"`
	UseCase            string         `json:"use_case,omitempty"`
	DataClassification string         `json:"data_classification"`
	RequestedModel     string         `json:"requested_model"`
	Prompt             string         `json:"prompt"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

type Route struct {
	RequestedModel string `json:"requested_model"`
	RoutedModel    string `json:"routed_model"`
	Provider       string `json:"provider"`
	Reason         string `json:"reason,omitempty"`
}

type Usage struct {
	Provider         string   `json:"provider"`
	Model            string   `json:"model"`
	InputTokens      int64    `json:"input_tokens"`
	OutputTokens     int64    `json:"output_tokens"`
	EstimatedCostUSD *float64 `json:"estimated_cost_usd"`
}

type ModelResponse struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Content  string `json:"content"`
}

type Result struct {
	Governance     governance.Request `json:"governance"`
	ProviderCalled bool               `json:"provider_called"`
	Route          *Route             `json:"route,omitempty"`
	Response       *ModelResponse     `json:"response,omitempty"`
	Usage          *Usage             `json:"usage,omitempty"`
}
