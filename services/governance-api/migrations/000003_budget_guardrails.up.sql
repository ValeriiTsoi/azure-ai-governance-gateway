CREATE TABLE budget_policies (
    id                          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    policy_name                 TEXT NOT NULL,
    cost_center                 TEXT NOT NULL,
    currency                    TEXT NOT NULL DEFAULT 'USD',

    monthly_limit_usd           NUMERIC(18, 8) NOT NULL,
    review_threshold_percent    NUMERIC(5, 2) NOT NULL DEFAULT 80.00,

    enabled                     BOOLEAN NOT NULL DEFAULT TRUE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT budget_policies_policy_name_check
        CHECK (btrim(policy_name) <> ''),

    CONSTRAINT budget_policies_cost_center_check
        CHECK (btrim(cost_center) <> ''),

    CONSTRAINT budget_policies_currency_check
        CHECK (currency = 'USD'),

    CONSTRAINT budget_policies_monthly_limit_check
        CHECK (monthly_limit_usd >= 0),

    CONSTRAINT budget_policies_review_threshold_check
        CHECK (
            review_threshold_percent >= 0
            AND review_threshold_percent <= 100
        ),

    CONSTRAINT budget_policies_cost_center_unique
        UNIQUE (cost_center)
);

CREATE TABLE budget_decisions (
    id                          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    governance_request_id       UUID NOT NULL
        REFERENCES governance_requests(id) ON DELETE CASCADE,

    budget_policy_id            BIGINT
        REFERENCES budget_policies(id) ON DELETE SET NULL,

    policy_name                 TEXT NOT NULL,
    cost_center                 TEXT,
    currency                    TEXT NOT NULL DEFAULT 'USD',

    decision                    TEXT NOT NULL,
    reason                      TEXT NOT NULL,

    period_start                TIMESTAMPTZ NOT NULL,
    period_end                  TIMESTAMPTZ NOT NULL,

    spent_usd                   NUMERIC(18, 8) NOT NULL,

    monthly_limit_usd           NUMERIC(18, 8),
    review_threshold_percent    NUMERIC(5, 2),
    utilization_percent         NUMERIC(9, 4),

    unknown_cost_records        BIGINT NOT NULL DEFAULT 0,

    evaluated_at                TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT budget_decisions_request_unique
        UNIQUE (governance_request_id),

    CONSTRAINT budget_decisions_policy_name_check
        CHECK (btrim(policy_name) <> ''),

    CONSTRAINT budget_decisions_currency_check
        CHECK (currency = 'USD'),

    CONSTRAINT budget_decisions_decision_check
        CHECK (decision IN ('allow', 'review', 'deny')),

    CONSTRAINT budget_decisions_period_check
        CHECK (period_end > period_start),

    CONSTRAINT budget_decisions_spent_check
        CHECK (spent_usd >= 0),

    CONSTRAINT budget_decisions_monthly_limit_check
        CHECK (
            monthly_limit_usd IS NULL
            OR monthly_limit_usd >= 0
        ),

    CONSTRAINT budget_decisions_review_threshold_check
        CHECK (
            review_threshold_percent IS NULL
            OR (
                review_threshold_percent >= 0
                AND review_threshold_percent <= 100
            )
        ),

    CONSTRAINT budget_decisions_utilization_check
        CHECK (
            utilization_percent IS NULL
            OR utilization_percent >= 0
        ),

    CONSTRAINT budget_decisions_unknown_cost_check
        CHECK (unknown_cost_records >= 0)
);

CREATE INDEX idx_budget_decisions_cost_center
    ON budget_decisions (cost_center, evaluated_at DESC);

CREATE INDEX idx_budget_decisions_evaluated_at
    ON budget_decisions (evaluated_at DESC);
