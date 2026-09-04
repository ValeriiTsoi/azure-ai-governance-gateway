package governance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	request Request,
	decision PolicyDecision,
) (Request, error) {
	metadata, err := json.Marshal(request.Metadata)
	if err != nil {
		return Request{}, fmt.Errorf("encode metadata: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Request{}, fmt.Errorf("begin governance transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var governanceID string

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO governance_requests (
			request_id,
			caller_subject,
			cost_center,
			use_case,
			data_classification,
			requested_model,
			prompt_hash,
			prompt_chars,
			metadata
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9::jsonb
		)
		RETURNING
			id::text,
			created_at
		`,
		request.RequestID,
		request.CallerSubject,
		nullableString(request.CostCenter),
		nullableString(request.UseCase),
		request.DataClassification,
		nullableString(request.RequestedModel),
		request.PromptHash,
		request.PromptChars,
		metadata,
	).Scan(
		&governanceID,
		&request.CreatedAt,
	)
	if err != nil {
		return Request{}, fmt.Errorf(
			"insert governance request: %w",
			err,
		)
	}

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO policy_decisions (
			governance_request_id,
			policy_name,
			decision,
			reason
		)
		VALUES ($1::uuid, $2, $3, $4)
		RETURNING evaluated_at
		`,
		governanceID,
		decision.PolicyName,
		decision.Decision,
		decision.Reason,
	).Scan(&decision.EvaluatedAt)
	if err != nil {
		return Request{}, fmt.Errorf(
			"insert policy decision: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return Request{}, fmt.Errorf(
			"commit governance transaction: %w",
			err,
		)
	}

	request.Policy = decision

	return request, nil
}

func (r *PostgresRepository) GetByRequestID(
	ctx context.Context,
	requestID string,
) (Request, error) {
	var request Request
	var metadata []byte

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			gr.request_id,
			gr.caller_subject,
			COALESCE(gr.cost_center, ''),
			COALESCE(gr.use_case, ''),
			gr.data_classification,
			COALESCE(gr.requested_model, ''),
			COALESCE(gr.prompt_hash, ''),
			COALESCE(gr.prompt_chars, 0),
			gr.metadata,
			gr.created_at,
			pd.policy_name,
			pd.decision,
			COALESCE(pd.reason, ''),
			pd.evaluated_at
		FROM governance_requests gr
		JOIN LATERAL (
			SELECT
				policy_name,
				decision,
				reason,
				evaluated_at
			FROM policy_decisions
			WHERE governance_request_id = gr.id
			ORDER BY evaluated_at DESC, id DESC
			LIMIT 1
		) pd ON true
		WHERE gr.request_id = $1
		`,
		requestID,
	).Scan(
		&request.RequestID,
		&request.CallerSubject,
		&request.CostCenter,
		&request.UseCase,
		&request.DataClassification,
		&request.RequestedModel,
		&request.PromptHash,
		&request.PromptChars,
		&metadata,
		&request.CreatedAt,
		&request.Policy.PolicyName,
		&request.Policy.Decision,
		&request.Policy.Reason,
		&request.Policy.EvaluatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrNotFound
	}

	if err != nil {
		return Request{}, fmt.Errorf(
			"query governance request: %w",
			err,
		)
	}

	if err := json.Unmarshal(metadata, &request.Metadata); err != nil {
		return Request{}, fmt.Errorf(
			"decode governance metadata: %w",
			err,
		)
	}

	return request, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	return value
}
