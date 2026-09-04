package governance

import "time"

type CreateRequestInput struct {
	CallerSubject      string         `json:"caller_subject"`
	CostCenter         string         `json:"cost_center,omitempty"`
	UseCase            string         `json:"use_case,omitempty"`
	DataClassification string         `json:"data_classification"`
	RequestedModel     string         `json:"requested_model"`
	Prompt             string         `json:"prompt"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

type PolicyDecision struct {
	PolicyName  string    `json:"policy_name"`
	Decision    string    `json:"decision"`
	Reason      string    `json:"reason"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

type Request struct {
	RequestID          string         `json:"request_id"`
	CallerSubject      string         `json:"caller_subject"`
	CostCenter         string         `json:"cost_center,omitempty"`
	UseCase            string         `json:"use_case,omitempty"`
	DataClassification string         `json:"data_classification"`
	RequestedModel     string         `json:"requested_model"`
	PromptHash         string         `json:"prompt_hash"`
	PromptChars        int            `json:"prompt_chars"`
	Metadata           map[string]any `json:"metadata"`
	CreatedAt          time.Time      `json:"created_at"`
	Policy             PolicyDecision `json:"policy"`
}
