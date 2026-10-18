package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var (
	ErrInvalidPaymentRequest = errors.New("invalid payment request")
	ErrPaymentOrderNotFound  = errors.New("order not found")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrPaymentConflict       = errors.New("payment state conflict")
	ErrIdempotencyConflict   = errors.New("idempotency key conflict")
	ErrInvalidPaymentOutcome = errors.New("invalid payment outcome")
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) CreateForOrder(
	ctx context.Context,
	userID, orderID, idempotencyKey string,
) (domain.Payment, error) {
	userID = strings.TrimSpace(userID)
	orderID = strings.TrimSpace(orderID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if userID == "" || orderID == "" || len(idempotencyKey) < 8 || len(idempotencyKey) > 200 {
		return domain.Payment{}, ErrInvalidPaymentRequest
	}
	p, err := s.repo.CreateOrGetPending(ctx, userID, orderID, idempotencyKey)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Payment{}, ErrPaymentOrderNotFound
	case errors.Is(err, repository.ErrPaymentIdempotencyConflict):
		return domain.Payment{}, ErrIdempotencyConflict
	case errors.Is(err, repository.ErrPaymentStateConflict):
		return domain.Payment{}, ErrPaymentConflict
	default:
		return p, err
	}
}

func (s *PaymentService) ProcessMockCallback(
	ctx context.Context,
	paymentID, outcome, eventID string,
) (domain.Payment, error) {
	paymentID = strings.TrimSpace(paymentID)
	outcome = strings.ToUpper(strings.TrimSpace(outcome))
	eventID = strings.TrimSpace(eventID)
	if paymentID == "" || len(eventID) < 3 || len(eventID) > 200 {
		return domain.Payment{}, ErrInvalidPaymentRequest
	}
	if outcome != "SUCCESS" && outcome != "FAILED" {
		return domain.Payment{}, ErrInvalidPaymentOutcome
	}
	p, err := s.repo.ProcessMockCallback(ctx, paymentID, outcome, eventID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Payment{}, ErrPaymentNotFound
	case errors.Is(err, repository.ErrPaymentIdempotencyConflict):
		return domain.Payment{}, ErrIdempotencyConflict
	case errors.Is(err, repository.ErrPaymentStateConflict):
		return domain.Payment{}, ErrPaymentConflict
	default:
		return p, err
	}
}

func (s *PaymentService) ReconciliationReport(
	ctx context.Context,
	olderThanMinutes int,
) (domain.PaymentReconciliationReport, error) {
	if olderThanMinutes < 1 || olderThanMinutes > 10080 {
		return domain.PaymentReconciliationReport{}, ErrInvalidPaymentRequest
	}
	return s.repo.ReconciliationReport(ctx, olderThanMinutes)
}
