package integration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

const integrationDatabaseName = "orderflow_test"

type testFixture struct {
	pool      *pgxpool.Pool
	ctx       context.Context
	userID    string
	productID string
	suffix    string
}

func openIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("ORDERFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set ORDERFLOW_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse ORDERFLOW_TEST_DATABASE_URL: %v", err)
	}
	if config.ConnConfig.Database != integrationDatabaseName {
		t.Fatalf("refusing to run integration tests against database %q; database must be exactly %q", config.ConnConfig.Database, integrationDatabaseName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect to dedicated integration database %q: %v", integrationDatabaseName, err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping dedicated integration database %q: %v", integrationDatabaseName, err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func newFixture(t *testing.T, initialStock int) *testFixture {
	t.Helper()
	pool := openIntegrationPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)

	suffix := randomSuffix(t)
	f := &testFixture{pool: pool, ctx: ctx, suffix: suffix}
	// Clean up only rows created by this fixture. The pool-close cleanup was
	// registered earlier, so this cleanup runs before closing the pool.
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if f.userID != "" {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM orders WHERE user_id = $1::uuid`, f.userID)
		}
		if f.productID != "" {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM inventory_movements WHERE product_id = $1::uuid`, f.productID)
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM inventory WHERE product_id = $1::uuid`, f.productID)
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM products WHERE id = $1::uuid`, f.productID)
		}
		if f.userID != "" {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id = $1::uuid`, f.userID)
		}
	})

	email := fmt.Sprintf("integration-%s@example.test", suffix)
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, 'integration-test-only', 'CUSTOMER')
		RETURNING id::text
	`, email).Scan(&f.userID)
	if err != nil {
		t.Fatalf("create integration test user: %v", err)
	}

	sku := "IT-" + suffix
	err = pool.QueryRow(ctx, `
		INSERT INTO products (sku, name, description, price_paise, is_active)
		VALUES ($1, 'Integration Test Product', 'Created by an isolated integration test', 49900, TRUE)
		RETURNING id::text
	`, sku).Scan(&f.productID)
	if err != nil {
		t.Fatalf("create integration test product: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO inventory (product_id, on_hand_quantity, reserved_quantity)
		VALUES ($1::uuid, $2, 0)
	`, f.productID, initialStock)
	if err != nil {
		t.Fatalf("create integration inventory: %v", err)
	}
	return f
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		t.Fatalf("generate fixture suffix: %v", err)
	}
	return hex.EncodeToString(bytes[:])
}

func (f *testFixture) createOrder(t *testing.T, quantity int) domain.CreateOrderResult {
	t.Helper()
	serviceUnderTest := service.NewOrderService(repository.NewOrderRepository(f.pool))
	result, err := serviceUnderTest.CreateOrder(f.ctx, domain.CreateOrderRequest{
		UserID: f.userID, ProductID: f.productID, Quantity: quantity,
	})
	if err != nil {
		t.Fatalf("create order through service: %v", err)
	}
	return result
}

func (f *testFixture) createPendingPayment(t *testing.T, order domain.CreateOrderResult) string {
	t.Helper()
	var paymentID string
	key := "integration-" + f.suffix
	err := f.pool.QueryRow(f.ctx, `
		INSERT INTO payments (order_id, provider, status, amount_paise, idempotency_key)
		VALUES ($1::uuid, 'MOCK', 'PENDING', $2, $3)
		RETURNING id::text
	`, order.OrderID, order.TotalAmountPaise, key).Scan(&paymentID)
	if err != nil {
		t.Fatalf("create pending payment fixture: %v", err)
	}
	return paymentID
}

func (f *testFixture) inventory(t *testing.T) (onHand, reserved int) {
	t.Helper()
	err := f.pool.QueryRow(f.ctx, `
		SELECT on_hand_quantity, reserved_quantity FROM inventory WHERE product_id = $1::uuid
	`, f.productID).Scan(&onHand, &reserved)
	if err != nil {
		t.Fatalf("read inventory fixture: %v", err)
	}
	return onHand, reserved
}

func (f *testFixture) orderStatus(t *testing.T, orderID string) string {
	t.Helper()
	var status string
	if err := f.pool.QueryRow(f.ctx, `SELECT status FROM orders WHERE id = $1::uuid`, orderID).Scan(&status); err != nil {
		t.Fatalf("read order status: %v", err)
	}
	return status
}

func (f *testFixture) movementCount(t *testing.T, orderID, movementType string) int {
	t.Helper()
	var count int
	if err := f.pool.QueryRow(f.ctx, `
		SELECT COUNT(*) FROM inventory_movements
		WHERE product_id = $1::uuid AND reference_type = 'ORDER'
		  AND reference_id = $2::uuid AND movement_type = $3
	`, f.productID, orderID, movementType).Scan(&count); err != nil {
		t.Fatalf("count inventory movements: %v", err)
	}
	return count
}

func TestOrderCreationReservesStockAndRecordsMovement(t *testing.T) {
	f := newFixture(t, 5)
	order := f.createOrder(t, 2)
	if order.Status != "PENDING" || order.TotalAmountPaise != 99800 {
		t.Fatalf("unexpected order result: %+v", order)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 5 || reserved != 2 {
		t.Fatalf("expected on_hand=5 reserved=2, got on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RESERVED"); got != 1 {
		t.Fatalf("expected one ORDER_RESERVED movement, got %d", got)
	}
	var itemCount int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM order_items WHERE order_id = $1::uuid`, order.OrderID).Scan(&itemCount); err != nil {
		t.Fatalf("count order items: %v", err)
	}
	if itemCount != 1 {
		t.Fatalf("expected one order item, got %d", itemCount)
	}
}

func TestInsufficientStockDoesNotCreatePartialOrder(t *testing.T) {
	f := newFixture(t, 1)
	svc := service.NewOrderService(repository.NewOrderRepository(f.pool))
	_, err := svc.CreateOrder(f.ctx, domain.CreateOrderRequest{
		UserID: f.userID, ProductID: f.productID, Quantity: 2,
	})
	if !errors.Is(err, service.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 1 || reserved != 0 {
		t.Fatalf("stock changed after rejected order: on_hand=%d reserved=%d", onHand, reserved)
	}
	var orderCount int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1::uuid`, f.userID).Scan(&orderCount); err != nil {
		t.Fatalf("count orders after failed request: %v", err)
	}
	if orderCount != 0 {
		t.Fatalf("expected no partially-created order, got %d", orderCount)
	}
}

func TestSuccessfulPaymentConfirmsOrderAndKeepsReservation(t *testing.T) {
	f := newFixture(t, 3)
	order := f.createOrder(t, 1)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	eventID := "success-" + f.suffix

	payment, err := payments.ProcessMockCallback(f.ctx, paymentID, "SUCCESS", eventID)
	if err != nil {
		t.Fatalf("process successful payment callback: %v", err)
	}
	if payment.Status != "SUCCESS" {
		t.Fatalf("expected SUCCESS payment, got %s", payment.Status)
	}
	if got := f.orderStatus(t, order.OrderID); got != "CONFIRMED" {
		t.Fatalf("expected CONFIRMED order, got %s", got)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 3 || reserved != 1 {
		t.Fatalf("successful payment must retain reservation: on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RELEASED"); got != 0 {
		t.Fatalf("successful payment should not release stock; got %d releases", got)
	}

	if _, err := payments.ProcessMockCallback(f.ctx, paymentID, "SUCCESS", eventID); err != nil {
		t.Fatalf("replay successful callback: %v", err)
	}
	onHand, reserved = f.inventory(t)
	if onHand != 3 || reserved != 1 {
		t.Fatalf("replayed success changed inventory: on_hand=%d reserved=%d", onHand, reserved)
	}
}

func TestFailedPaymentReleasesReservationExactlyOnce(t *testing.T) {
	f := newFixture(t, 7)
	order := f.createOrder(t, 2)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	eventID := "failure-" + f.suffix

	payment, err := payments.ProcessMockCallback(f.ctx, paymentID, "FAILED", eventID)
	if err != nil {
		t.Fatalf("process failed payment callback: %v", err)
	}
	if payment.Status != "FAILED" {
		t.Fatalf("expected FAILED payment, got %s", payment.Status)
	}
	if got := f.orderStatus(t, order.OrderID); got != "FAILED" {
		t.Fatalf("expected FAILED order, got %s", got)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 7 || reserved != 0 {
		t.Fatalf("failed payment did not release reservation: on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RELEASED"); got != 1 {
		t.Fatalf("expected one release movement, got %d", got)
	}

	if _, err := payments.ProcessMockCallback(f.ctx, paymentID, "FAILED", eventID); err != nil {
		t.Fatalf("replay failed callback: %v", err)
	}
	onHand, reserved = f.inventory(t)
	if onHand != 7 || reserved != 0 {
		t.Fatalf("replayed failure changed inventory: on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RELEASED"); got != 1 {
		t.Fatalf("expected exactly one release movement after replay, got %d", got)
	}
}

func TestPaymentInitiationIsIdempotentForSameKey(t *testing.T) {
	f := newFixture(t, 2)
	order := f.createOrder(t, 1)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	key := "integration-key-" + f.suffix

	first, err := payments.CreateForOrder(f.ctx, f.userID, order.OrderID, key)
	if err != nil {
		t.Fatalf("create first payment: %v", err)
	}
	second, err := payments.CreateForOrder(f.ctx, f.userID, order.OrderID, key)
	if err != nil {
		t.Fatalf("repeat payment with same idempotency key: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("same idempotency key returned different IDs: first=%s second=%s", first.ID, second.ID)
	}
	var count int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM payments WHERE order_id = $1::uuid`, order.OrderID).Scan(&count); err != nil {
		t.Fatalf("count payment attempts: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one payment record, got %d", count)
	}
}
