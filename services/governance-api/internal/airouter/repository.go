package airouter

import (
	"context"
	"fmt"

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

	usageResult, err := tx.Exec(
		ctx,
		`
		INSERT INTO usage_records (
			governance_request_id,
			provider,
			model,
			input_tokens,
			output_tokens,
			estimated_cost_usd
		)
		SELECT
			id,
			$2,
			$3,
			$4,
			$5,
			$6
		FROM governance_requests
		WHERE request_id = $1
		`,
		requestID,
		usage.Provider,
		usage.Model,
		usage.InputTokens,
		usage.OutputTokens,
		usage.EstimatedCostUSD,
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
