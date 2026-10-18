package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

var ErrMockProviderOutcomeAlreadyFinalized = errors.New("mock provider outcome is already final")

func scanMockProviderOutcome(row pgx.Row) (domain.MockProviderOutcome, error) {
	var outcome domain.MockProviderOutcome
	err := row.Scan(
		&outcome.PaymentID,
		&outcome.ProviderPaymentID,
		&outcome.Outcome,
		&outcome.AmountPaise,
		&outcome.Currency,
		&outcome.FailureReason,
		&outcome.UpdatedAt,
	)
	return outcome, err
}

// SetMockProviderOutcome changes only the simulated provider's record. It does
// not update the application's payments/orders; reconciliation must do that.
func (r *PaymentRepository) SetMockProviderOutcome(
	ctx context.Context,
	paymentID, outcome string,
	amountPaise int64,
	currency, failureReason string,
) (domain.MockProviderOutcome, error) {
	outcome = strings.ToUpper(strings.TrimSpace(outcome))
	currency = strings.ToUpper(strings.TrimSpace(currency))
	failureReason = strings.TrimSpace(failureReason)
	result, err := scanMockProviderOutcome(r.pool.QueryRow(ctx, `
		UPDATE mock_provider_transactions
		SET status = $2,
		    amount_paise = $3,
		    currency = $4,
		    failure_reason = NULLIF($5, ''),
		    updated_at = NOW()
		WHERE payment_id = $1::uuid AND status = 'PENDING'
		RETURNING payment_id::text, provider_payment_id, status, amount_paise,
		          currency, COALESCE(failure_reason, ''), updated_at
	`, paymentID, outcome, amountPaise, currency, failureReason))
	if err == nil {
		return result, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.MockProviderOutcome{}, err
	}
	var currentStatus string
	lookupErr := r.pool.QueryRow(ctx, `
		SELECT status FROM mock_provider_transactions WHERE payment_id = $1::uuid
	`, paymentID).Scan(&currentStatus)
	if errors.Is(lookupErr, pgx.ErrNoRows) {
		return domain.MockProviderOutcome{}, pgx.ErrNoRows
	}
	if lookupErr != nil {
		return domain.MockProviderOutcome{}, lookupErr
	}
	return domain.MockProviderOutcome{}, ErrMockProviderOutcomeAlreadyFinalized
}

// ListMockProviderOutcomes returns provider-side terminal results whose
// corresponding OrderFlow payment is still pending and old enough to reconcile.
func (r *PaymentRepository) ListMockProviderOutcomes(
	ctx context.Context,
	olderThanMinutes, limit int,
) ([]domain.MockProviderOutcome, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.payment_id::text, m.provider_payment_id, m.status,
		       m.amount_paise, m.currency, COALESCE(m.failure_reason, ''), m.updated_at
		FROM mock_provider_transactions m
		JOIN payments p ON p.id = m.payment_id
		WHERE p.status = 'PENDING'
		  AND p.provider = 'MOCK'
		  AND m.status IN ('SUCCESS', 'FAILED')
		  AND m.updated_at <= NOW() - ($1::int * INTERVAL '1 minute')
		ORDER BY m.updated_at ASC, m.payment_id ASC
		LIMIT $2
	`, olderThanMinutes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	outcomes := make([]domain.MockProviderOutcome, 0)
	for rows.Next() {
		var outcome domain.MockProviderOutcome
		if err := rows.Scan(
			&outcome.PaymentID, &outcome.ProviderPaymentID, &outcome.Outcome,
			&outcome.AmountPaise, &outcome.Currency, &outcome.FailureReason,
			&outcome.UpdatedAt,
		); err != nil {
			return nil, err
		}
		outcomes = append(outcomes, outcome)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return outcomes, nil
}
