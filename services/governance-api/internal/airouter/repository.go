package airouter

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(
	db *pgxpool.Pool,
) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) RecordInvocation(
	ctx context.Context,
	requestID string,
	route Route,
	usage Usage,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin AI invocation transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	routeResult, err := tx.Exec(
		ctx,
		`
		INSERT INTO model_routes (
			governance_request_id,
			requested_model,
			routed_model,
			provider,
			reason
		)
		SELECT
			id,
			$2,
			$3,
			$4,
			$5
		FROM governance_requests
		WHERE request_id = $1
		`,
		requestID,
		route.RequestedModel,
		route.RoutedModel,
		route.Provider,
		route.Reason,
	)
	if err != nil {
		return fmt.Errorf(
			"insert model route: %w",
			err,
		)
	}

	if routeResult.RowsAffected() != 1 {
		return fmt.Errorf(
			"governance request %q not found while recording model route",
			requestID,
		)
	}

	var (
		pricingSource         *string
		pricingEffectiveStart *time.Time
		inputPrice            *float64
		cachedInputPrice      *float64
		outputPrice           *float64
	)

	if usage.Pricing != nil {
		if usage.Pricing.Source != "" {
			value := usage.Pricing.Source
			pricingSource = &value
		}

		if usage.Pricing.EffectiveStartDate != "" {
			value, err := time.Parse(
				time.RFC3339,
				usage.Pricing.EffectiveStartDate,
			)
			if err != nil {
				return fmt.Errorf(
					"parse pricing effective start date: %w",
					err,
				)
			}

			pricingEffectiveStart = &value
		}

		value := usage.Pricing.InputPerMillionUSD
		inputPrice = &value

		value = usage.Pricing.CachedInputPerMillionUSD
		cachedInputPrice = &value

		value = usage.Pricing.OutputPerMillionUSD
		outputPrice = &value
	}

	usageResult, err := tx.Exec(
		ctx,
		`
		INSERT INTO usage_records (
			governance_request_id,
			provider,
			model,
			input_tokens,
			cached_input_tokens,
			output_tokens,
			estimated_cost_usd,
			pricing_source,
			pricing_effective_start_date,
			input_price_per_million_usd,
			cached_input_price_per_million_usd,
			output_price_per_million_usd
		)
		SELECT
			id,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12
		FROM governance_requests
		WHERE request_id = $1
		`,
		requestID,
		usage.Provider,
		usage.Model,
		usage.InputTokens,
		usage.CachedInputTokens,
		usage.OutputTokens,
		usage.EstimatedCostUSD,
		pricingSource,
		pricingEffectiveStart,
		inputPrice,
		cachedInputPrice,
		outputPrice,
	)
	if err != nil {
		return fmt.Errorf(
			"insert usage record: %w",
			err,
		)
	}

	if usageResult.RowsAffected() != 1 {
		return fmt.Errorf(
			"governance request %q not found while recording usage",
			requestID,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit AI invocation transaction: %w",
			err,
		)
	}

	return nil
}
