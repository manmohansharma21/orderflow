package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestCreateForOrderRejectsInvalidInput(t *testing.T) {
	svc := NewPaymentService(nil)
	validOrderID := "8e840ec0-9315-4574-847f-15e8a17a578a"

	tests := []struct {
		name           string
		userID         string
		orderID        string
		idempotencyKey string
	}{
		{name: "missing user", userID: "", orderID: validOrderID, idempotencyKey: "key-123456"},
		{name: "missing order", userID: "user-1", orderID: "", idempotencyKey: "key-123456"},
		{name: "short idempotency key", userID: "user-1", orderID: validOrderID, idempotencyKey: "short"},
		{name: "long idempotency key", userID: "user-1", orderID: validOrderID, idempotencyKey: strings.Repeat("k", 201)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateForOrder(context.Background(), tt.userID, tt.orderID, tt.idempotencyKey)
			if !errors.Is(err, ErrInvalidPaymentRequest) {
				t.Fatalf("expected ErrInvalidPaymentRequest, got %v", err)
			}
		})
	}
}

func TestProcessMockCallbackRejectsInvalidInput(t *testing.T) {
	svc := NewPaymentService(nil)
	validPaymentID := "28e3f2b2-cfc8-4df9-8b6c-d0f7104c9b1d"

	tests := []struct {
		name      string
		paymentID string
		outcome   string
		eventID   string
		wantErr   error
	}{
		{name: "missing payment ID", paymentID: "", outcome: "SUCCESS", eventID: "evt-123", wantErr: ErrInvalidPaymentRequest},
		{name: "missing event ID", paymentID: validPaymentID, outcome: "SUCCESS", eventID: "", wantErr: ErrInvalidPaymentRequest},
		{name: "short event ID", paymentID: validPaymentID, outcome: "FAILED", eventID: "ab", wantErr: ErrInvalidPaymentRequest},
		{name: "long event ID", paymentID: validPaymentID, outcome: "FAILED", eventID: strings.Repeat("e", 201), wantErr: ErrInvalidPaymentRequest},
		{name: "unsupported outcome", paymentID: validPaymentID, outcome: "REFUNDED", eventID: "evt-123", wantErr: ErrInvalidPaymentOutcome},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ProcessMockCallback(context.Background(), tt.paymentID, tt.outcome, tt.eventID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestReconciliationReportRejectsInvalidThreshold(t *testing.T) {
	svc := NewPaymentService(nil)
	for _, minutes := range []int{-1, 0, 10081} {
		t.Run(fmt.Sprintf("threshold_%d_minutes", minutes), func(t *testing.T) {
			_, err := svc.ReconciliationReport(context.Background(), minutes)
			if !errors.Is(err, ErrInvalidPaymentRequest) {
				t.Fatalf("expected ErrInvalidPaymentRequest for %d minutes, got %v", minutes, err)
			}
		})
	}
}
