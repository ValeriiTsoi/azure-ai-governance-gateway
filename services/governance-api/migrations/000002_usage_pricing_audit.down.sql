ALTER TABLE usage_records
    DROP CONSTRAINT IF EXISTS usage_records_output_price_check,
    DROP CONSTRAINT IF EXISTS usage_records_cached_input_price_check,
    DROP CONSTRAINT IF EXISTS usage_records_input_price_check,
    DROP CONSTRAINT IF EXISTS usage_records_cached_input_within_input_check,
    DROP CONSTRAINT IF EXISTS usage_records_cached_input_tokens_check;

ALTER TABLE usage_records
    DROP COLUMN IF EXISTS output_price_per_million_usd,
    DROP COLUMN IF EXISTS cached_input_price_per_million_usd,
    DROP COLUMN IF EXISTS input_price_per_million_usd,
    DROP COLUMN IF EXISTS pricing_effective_start_date,
    DROP COLUMN IF EXISTS pricing_source,
    DROP COLUMN IF EXISTS cached_input_tokens;
