package budget

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindActivePolicy(
		context.Context,
		string,
	) (Policy, bool, error)

	CurrentSpend(
		context.Context,
		string,
		time.Time,
		time.Time,
	) (SpendSnapshot, error)

	RecordDecision(
		context.Context,
		string,
		Decision,
	) error
}

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

func (r *PostgresRepository) FindActivePolicy(
	ctx context.Context,
	costCenter string,
) (Policy, bool, error) {
	var policy Policy

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			policy_name,
			cost_center,
			currency,
			monthly_limit_usd,
			review_threshold_percent
		FROM budget_policies
		WHERE cost_center = $1
		  AND enabled = true
		LIMIT 1
		`,
		strings.TrimSpace(costCenter),
	).Scan(
		&policy.ID,
		&policy.PolicyName,
		&policy.CostCenter,
		&policy.Currency,
		&policy.MonthlyLimitUSD,
		&policy.ReviewThresholdPercent,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Policy{}, false, nil
	}

	if err != nil {
		return Policy{}, false, fmt.Errorf(
			"query active budget policy: %w",
			err,
		)
	}

	return policy, true, nil
}

func (r *PostgresRepository) CurrentSpend(
	ctx context.Context,
	costCenter string,
	periodStart time.Time,
	periodEnd time.Time,
) (SpendSnapshot, error) {
	var snapshot SpendSnapshot

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			COALESCE(
				SUM(ur.estimated_cost_usd)
					FILTER (
						WHERE ur.estimated_cost_usd IS NOT NULL
					),
				0
			)::double precision,
			COUNT(*)
				FILTER (
					WHERE ur.estimated_cost_usd IS NULL
				)
		FROM usage_records ur
		JOIN governance_requests gr
		  ON gr.id = ur.governance_request_id
		WHERE gr.cost_center = $1
		  AND ur.recorded_at >= $2
		  AND ur.recorded_at < $3
		`,
		strings.TrimSpace(costCenter),
		periodStart,
		periodEnd,
	).Scan(
		&snapshot.SpentUSD,
		&snapshot.UnknownCostRecords,
	)

	if err != nil {
		return SpendSnapshot{}, fmt.Errorf(
			"query current budget spend: %w",
			err,
		)
	}

	return snapshot, nil
}

func (r *PostgresRepository) RecordDecision(
	ctx context.Context,
	requestID string,
	decision Decision,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		INSERT INTO budget_decisions (
			governance_request_id,
			budget_policy_id,
			policy_name,
			cost_center,
			currency,
			decision,
			reason,
			period_start,
			period_end,
			spent_usd,
			monthly_limit_usd,
			review_threshold_percent,
			utilization_percent,
			unknown_cost_records,
			evaluated_at
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
			$12,
			$13,
			$14,
			$15
		FROM governance_requests
		WHERE request_id = $1
		`,
		strings.TrimSpace(requestID),
		nullableInt64(decision.PolicyID),
		decision.PolicyName,
		nullableString(decision.CostCenter),
		decision.Currency,
		decision.Decision,
		decision.Reason,
		decision.PeriodStart,
		decision.PeriodEnd,
		decision.SpentUSD,
		nullableFloat64(decision.MonthlyLimitUSD),
		nullableFloat64(decision.ReviewThresholdPercent),
		nullableFloat64(decision.UtilizationPercent),
		decision.UnknownCostRecords,
		decision.EvaluatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"insert budget decision: %w",
			err,
		)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf(
			"governance request %q not found while recording budget decision",
			requestID,
		)
	}

	return nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	return value
}

func nullableFloat64(value *float64) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}

	return *value
}
