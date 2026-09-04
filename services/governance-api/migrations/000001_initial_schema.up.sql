CREATE TABLE governance_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id          TEXT NOT NULL UNIQUE,
    caller_subject      TEXT NOT NULL,
    cost_center         TEXT,
    use_case            TEXT,
    data_classification TEXT NOT NULL DEFAULT 'internal',
    requested_model     TEXT,
    prompt_hash         TEXT,
    prompt_chars        INTEGER,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT governance_requests_data_classification_check
        CHECK (
            data_classification IN (
                'public',
                'internal',
                'confidential',
                'restricted'
            )
        ),

    CONSTRAINT governance_requests_prompt_chars_check
        CHECK (prompt_chars IS NULL OR prompt_chars >= 0)
);

CREATE INDEX idx_governance_requests_created_at
    ON governance_requests (created_at DESC);

CREATE INDEX idx_governance_requests_caller_subject
    ON governance_requests (caller_subject);

CREATE INDEX idx_governance_requests_cost_center
    ON governance_requests (cost_center);


CREATE TABLE policy_decisions (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    governance_request_id UUID NOT NULL
        REFERENCES governance_requests(id) ON DELETE CASCADE,

    policy_name           TEXT NOT NULL,
    decision              TEXT NOT NULL,
    reason                TEXT,
    details               JSONB NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT policy_decisions_decision_check
        CHECK (
            decision IN (
                'allow',
                'deny',
                'review'
            )
        )
);

CREATE INDEX idx_policy_decisions_request
    ON policy_decisions (governance_request_id);


CREATE TABLE model_routes (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    governance_request_id UUID NOT NULL
        REFERENCES governance_requests(id) ON DELETE CASCADE,

    requested_model       TEXT,
    routed_model          TEXT NOT NULL,
    provider              TEXT NOT NULL,
    reason                TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_model_routes_request
    ON model_routes (governance_request_id);


CREATE TABLE usage_records (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    governance_request_id UUID NOT NULL
        REFERENCES governance_requests(id) ON DELETE CASCADE,

    provider              TEXT NOT NULL,
    model                 TEXT NOT NULL,
    input_tokens          BIGINT NOT NULL DEFAULT 0,
    output_tokens         BIGINT NOT NULL DEFAULT 0,
    estimated_cost_usd    NUMERIC(18, 8),
    recorded_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT usage_records_input_tokens_check
        CHECK (input_tokens >= 0),

    CONSTRAINT usage_records_output_tokens_check
        CHECK (output_tokens >= 0),

    CONSTRAINT usage_records_cost_check
        CHECK (estimated_cost_usd IS NULL OR estimated_cost_usd >= 0)
);

CREATE INDEX idx_usage_records_request
    ON usage_records (governance_request_id);

CREATE INDEX idx_usage_records_recorded_at
    ON usage_records (recorded_at DESC);
