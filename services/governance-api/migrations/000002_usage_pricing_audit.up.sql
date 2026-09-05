ALTER TABLE usage_records
    ADD COLUMN cached_input_tokens BIGINT,
    ADD COLUMN pricing_source TEXT,
    ADD COLUMN pricing_effective_start_date TIMESTAMPTZ,
    ADD COLUMN input_price_per_million_usd NUMERIC(18, 8),
    ADD COLUMN cached_input_price_per_million_usd NUMERIC(18, 8),
    ADD COLUMN output_price_per_million_usd NUMERIC(18, 8);

ALTER TABLE usage_records
    ADD CONSTRAINT usage_records_cached_input_tokens_check
        CHECK (
            cached_input_tokens IS NULL
            OR cached_input_tokens >= 0
        ),

    ADD CONSTRAINT usage_records_cached_input_within_input_check
        CHECK (
            cached_input_tokens IS NULL
            OR cached_input_tokens <= input_tokens
        ),

    ADD CONSTRAINT usage_records_input_price_check
        CHECK (
            input_price_per_million_usd IS NULL
            OR input_price_per_million_usd >= 0
        ),

    ADD CONSTRAINT usage_records_cached_input_price_check
        CHECK (
            cached_input_price_per_million_usd IS NULL
            OR cached_input_price_per_million_usd >= 0
        ),

    ADD CONSTRAINT usage_records_output_price_check
        CHECK (
            output_price_per_million_usd IS NULL
            OR output_price_per_million_usd >= 0
        );
