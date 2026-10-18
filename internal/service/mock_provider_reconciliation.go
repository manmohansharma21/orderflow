package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var ErrMockProviderOutcomeFinalized = errors.New("mock provider outcome is already final")

// SetMockProviderOutcome simulates the provider settling a payment without
// delivering its webhook. It is a development/demo-only control, not a real
// payment gateway and not a customer-facing operation.
func (s *PaymentService) SetMockProviderOutcome(
	ctx context.Context,
	paymentID, outcome string,
	amountPaise int64,
	currency, failureReason string,
) (domain.MockProviderOutcome, error) {
	paymentID = strings.TrimSpace(paymentID)
	outcome = strings.ToUpper(strings.TrimSpace(outcome))
	currency = strings.ToUpper(strings.TrimSpace(currency))
	failureReason = strings.TrimSpace(failureReason)
	if paymentID == "" || (outcome != "SUCCESS" && outcome != "FAILED") ||
		amountPaise < 0 || len(currency) != 3 ||
		(outcome == "FAILED" && len(failureReason) < 3) {
		return domain.MockProviderOutcome{}, ErrInvalidPaymentRequest
	}
	result, err := s.repo.SetMockProviderOutcome(ctx, paymentID, outcome, amountPaise, currency, failureReason)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MockProviderOutcome{}, ErrPaymentNotFound
	}
	if errors.Is(err, repository.ErrMockProviderOutcomeAlreadyFinalized) {
		return domain.MockProviderOutcome{}, ErrMockProviderOutcomeFinalized
	}
	return result, err
}

// ReconcileKnownMockProviderOutcomes polls the independent mock-provider ledger
// for known terminal outcomes and feeds them through ProcessMockCallback. This
// intentionally reuses the exact callback transition path for amount/currency
// validation, idempotency, order-state protection, stock updates and audit logs.
func (s *PaymentService) ReconcileKnownMockProviderOutcomes(
	ctx context.Context,
	olderThanMinutes, batchSize int,
) (int, error) {
	if olderThanMinutes < 0 || olderThanMinutes > 10080 || batchSize < 1 || batchSize > 1000 {
		return 0, ErrInvalidPaymentRequest
	}
	outcomes, err := s.repo.ListMockProviderOutcomes(ctx, olderThanMinutes, batchSize)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var firstError error
	for _, outcome := range outcomes {
		eventID := "reconcile-" + outcome.ProviderPaymentID
		_, err := s.ProcessMockCallback(ctx, outcome.PaymentID, outcome.Outcome, eventID, domain.PaymentCallbackDetails{
			AmountPaise:   outcome.AmountPaise,
			Currency:      outcome.Currency,
			FailureReason: outcome.FailureReason,
		})
		if err != nil {
			// A concurrent callback/cancellation may have completed the transition.
			// State conflicts are safe to skip; record other errors for visibility.
			if errors.Is(err, ErrPaymentConflict) {
				continue
			}
			if firstError == nil {
				firstError = err
			}
			continue
		}
		recovered++
	}
	return recovered, firstError
}
