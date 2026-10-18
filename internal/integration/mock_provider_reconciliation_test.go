package integration

import (
	"testing"
	"time"

	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

func TestReconciliationRecoversMissedMockProviderSuccess(t *testing.T) {
	f := newFixture(t, 6)
	order := f.createOrder(t, 2)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))

	if _, err := payments.SetMockProviderOutcome(f.ctx, paymentID, "SUCCESS", order.TotalAmountPaise, "INR", ""); err != nil {
		t.Fatalf("simulate provider success without callback: %v", err)
	}
	if _, err := f.pool.Exec(f.ctx, `UPDATE mock_provider_transactions SET updated_at=$2 WHERE payment_id=$1::uuid`, paymentID, time.Now().Add(-2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	count, err := payments.ReconcileKnownMockProviderOutcomes(f.ctx, 1, 10)
	if err != nil {
		t.Fatalf("reconcile missed success callback: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one recovered payment, got %d", count)
	}
	if got := f.orderStatus(t, order.OrderID); got != "CONFIRMED" {
		t.Fatalf("expected order CONFIRMED, got %s", got)
	}
	if got := mockPaymentStatus(t, f, paymentID); got != "SUCCESS" {
		t.Fatalf("expected payment SUCCESS, got %s", got)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 4 || reserved != 0 {
		t.Fatalf("successful reconciliation should consume stock, got on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_FULFILLED"); got != 1 {
		t.Fatalf("expected one fulfillment movement, got %d", got)
	}
	count, err = payments.ReconcileKnownMockProviderOutcomes(f.ctx, 0, 10)
	if err != nil || count != 0 {
		t.Fatalf("replay should not perform a second transition: count=%d err=%v", count, err)
	}
}

func TestReconciliationRecoversMissedMockProviderFailure(t *testing.T) {
	f := newFixture(t, 6)
	order := f.createOrder(t, 3)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))

	if _, err := payments.SetMockProviderOutcome(f.ctx, paymentID, "FAILED", order.TotalAmountPaise, "INR", "provider_declined"); err != nil {
		t.Fatalf("simulate provider failure without callback: %v", err)
	}
	if _, err := f.pool.Exec(f.ctx, `UPDATE mock_provider_transactions SET updated_at=$2 WHERE payment_id=$1::uuid`, paymentID, time.Now().Add(-2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	count, err := payments.ReconcileKnownMockProviderOutcomes(f.ctx, 1, 10)
	if err != nil {
		t.Fatalf("reconcile missed failure callback: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one recovered payment, got %d", count)
	}
	if got := f.orderStatus(t, order.OrderID); got != "FAILED" {
		t.Fatalf("expected order FAILED, got %s", got)
	}
	if got := mockPaymentStatus(t, f, paymentID); got != "FAILED" {
		t.Fatalf("expected payment FAILED, got %s", got)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 6 || reserved != 0 {
		t.Fatalf("failed payment should release reservation only, got on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RELEASED"); got != 1 {
		t.Fatalf("expected one release movement, got %d", got)
	}
}

func TestReconciliationDoesNotApplyMismatchedProviderAmount(t *testing.T) {
	f := newFixture(t, 5)
	order := f.createOrder(t, 1)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	if _, err := payments.SetMockProviderOutcome(f.ctx, paymentID, "SUCCESS", order.TotalAmountPaise+1, "INR", ""); err != nil {
		t.Fatalf("set synthetic mismatched provider outcome: %v", err)
	}
	if _, err := f.pool.Exec(f.ctx, `UPDATE mock_provider_transactions SET updated_at=$2 WHERE payment_id=$1::uuid`, paymentID, time.Now().Add(-2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := payments.ReconcileKnownMockProviderOutcomes(f.ctx, 1, 10); err == nil {
		t.Fatal("expected amount mismatch to be surfaced")
	}
	if got := mockPaymentStatus(t, f, paymentID); got != "PENDING" {
		t.Fatalf("mismatched outcome must not settle payment, got %s", got)
	}
	if got := f.orderStatus(t, order.OrderID); got != "PENDING" {
		t.Fatalf("mismatched outcome must not confirm order, got %s", got)
	}
}

func mockPaymentStatus(t *testing.T, f *testFixture, paymentID string) string {
	t.Helper()
	var status string
	if err := f.pool.QueryRow(f.ctx, `SELECT status FROM payments WHERE id=$1::uuid`, paymentID).Scan(&status); err != nil {
		t.Fatalf("read payment status: %v", err)
	}
	return status
}
