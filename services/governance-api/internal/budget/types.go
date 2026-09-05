package budget

import "time"

const PolicyName = "monthly-cost-center-budget-v1"

type Policy struct {
	ID                     int64
	PolicyName             string
	CostCenter             string
	Currency               string
	MonthlyLimitUSD        float64
	ReviewThresholdPercent float64
}

type SpendSnapshot struct {
	SpentUSD           float64
	UnknownCostRecords int64
}

type EvaluateInput struct {
	RequestID  string
	CostCenter string
}

type Decision struct {
	PolicyID *int64 `json:"-"`

	PolicyName string `json:"policy_name"`
	Decision   string `json:"decision"`
	Reason     string `json:"reason"`

	CostCenter string `json:"cost_center,omitempty"`
	Currency   string `json:"currency"`

	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	SpentUSD float64 `json:"spent_usd"`

	MonthlyLimitUSD        *float64 `json:"monthly_limit_usd,omitempty"`
	ReviewThresholdPercent *float64 `json:"review_threshold_percent,omitempty"`
	UtilizationPercent     *float64 `json:"utilization_percent,omitempty"`

	UnknownCostRecords int64     `json:"unknown_cost_records"`
	EvaluatedAt        time.Time `json:"evaluated_at"`
}
