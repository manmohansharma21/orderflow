package integration

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

func TestMultiProductOrderAndIdempotentReplay(t *testing.T) {
	f := newFixture(t, 8)
	secondSuffix := randomSuffix(t)
	var secondProduct string
	if err := f.pool.QueryRow(f.ctx, `INSERT INTO products (sku,name,price_paise,is_active) VALUES ($1,'Second Integration Product',12000,TRUE) RETURNING id::text`, "IT2-"+secondSuffix).Scan(&secondProduct); err != nil {
		t.Fatalf("create second product: %v", err)
	}
	if _, err := f.pool.Exec(f.ctx, `INSERT INTO inventory (product_id,on_hand_quantity,reserved_quantity) VALUES ($1::uuid,5,0)`, secondProduct); err != nil {
		t.Fatalf("create second inventory: %v", err)
	}
	t.Cleanup(func() {
		_, _ = f.pool.Exec(f.ctx, `DELETE FROM orders WHERE user_id=$1::uuid`, f.userID)
		_, _ = f.pool.Exec(f.ctx, `DELETE FROM inventory_movements WHERE product_id=$1::uuid`, secondProduct)
		_, _ = f.pool.Exec(f.ctx, `DELETE FROM inventory WHERE product_id=$1::uuid`, secondProduct)
		_, _ = f.pool.Exec(f.ctx, `DELETE FROM products WHERE id=$1::uuid`, secondProduct)
	})

	svc := service.NewOrderService(repository.NewOrderRepository(f.pool))
	req := domain.CreateOrderRequest{UserID: f.userID, IdempotencyKey: "multi-item-key-" + f.suffix, Items: []domain.CreateOrderLineRequest{{ProductID: f.productID, Quantity: 2}, {ProductID: secondProduct, Quantity: 1}}}
	first, err := svc.CreateOrder(f.ctx, req)
	if err != nil {
		t.Fatalf("create multi-product order: %v", err)
	}
	if first.TotalAmountPaise != 111800 {
		t.Fatalf("expected total 111800 paise, got %d", first.TotalAmountPaise)
	}
	replay, err := svc.CreateOrder(f.ctx, req)
	if err != nil {
		t.Fatalf("replay idempotent order: %v", err)
	}
	if replay.OrderID != first.OrderID || !replay.Replayed {
		t.Fatalf("same key did not replay original order: first=%+v replay=%+v", first, replay)
	}
	var itemCount int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM order_items WHERE order_id=$1::uuid`, first.OrderID).Scan(&itemCount); err != nil {
		t.Fatal(err)
	}
	if itemCount != 2 {
		t.Fatalf("expected 2 order item snapshots, got %d", itemCount)
	}
	changed := req
	changed.Items = []domain.CreateOrderLineRequest{{ProductID: f.productID, Quantity: 1}, {ProductID: secondProduct, Quantity: 1}}
	if _, err := svc.CreateOrder(f.ctx, changed); !errors.Is(err, service.ErrOrderIdempotencyConflict) {
		t.Fatalf("same key with changed payload should conflict; got %v", err)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 8 || reserved != 2 {
		t.Fatalf("idempotency replay duplicated reservation: on_hand=%d reserved=%d", onHand, reserved)
	}
}

func TestCallbackRejectsAmountOrCurrencyMismatchAndStoresFailureReason(t *testing.T) {
	f := newFixture(t, 4)
	order := f.createOrder(t, 1)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	if _, err := payments.ProcessMockCallback(f.ctx, paymentID, "SUCCESS", "amt-mismatch-"+f.suffix, domain.PaymentCallbackDetails{AmountPaise: order.TotalAmountPaise + 1, Currency: "INR"}); !errors.Is(err, service.ErrPaymentAmountMismatch) {
		t.Fatalf("expected amount mismatch error, got %v", err)
	}
	if _, err := payments.ProcessMockCallback(f.ctx, paymentID, "SUCCESS", "currency-mismatch-"+f.suffix, domain.PaymentCallbackDetails{AmountPaise: order.TotalAmountPaise, Currency: "USD"}); !errors.Is(err, service.ErrPaymentAmountMismatch) {
		t.Fatalf("expected currency mismatch error, got %v", err)
	}
	if got := f.orderStatus(t, order.OrderID); got != "PENDING" {
		t.Fatalf("mismatched callback changed order status: %s", got)
	}
	payment, err := payments.ProcessMockCallback(f.ctx, paymentID, "FAILED", "failed-reason-"+f.suffix, domain.PaymentCallbackDetails{AmountPaise: order.TotalAmountPaise, Currency: "INR", FailureReason: "provider_declined"})
	if err != nil {
		t.Fatalf("process reasoned failure callback: %v", err)
	}
	if payment.FailureReason != "provider_declined" {
		t.Fatalf("failure reason not returned/stored: %+v", payment)
	}
}

func TestExpiryReleasesReservationForUnpaidOrder(t *testing.T) {
	f := newFixture(t, 5)
	order := f.createOrder(t, 2)
	if _, err := f.pool.Exec(f.ctx, `UPDATE orders SET created_at=NOW()-INTERVAL '2 hours' WHERE id=$1::uuid`, order.OrderID); err != nil {
		t.Fatal(err)
	}
	svc := service.NewOrderService(repository.NewOrderRepository(f.pool))
	count, err := svc.ExpireUnpaidOrders(f.ctx, 60, 10)
	if err != nil {
		t.Fatalf("expire unpaid order: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one expired order, got %d", count)
	}
	var status string
	if err := f.pool.QueryRow(f.ctx, `SELECT status FROM orders WHERE id=$1::uuid`, order.OrderID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "EXPIRED" {
		t.Fatalf("expected EXPIRED status, got %s", status)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 5 || reserved != 0 {
		t.Fatalf("expiry did not release stock: on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RELEASED"); got != 1 {
		t.Fatalf("expected one expiry release movement, got %d", got)
	}
}

func TestExpiryDoesNotExpireOrderWithPendingPayment(t *testing.T) {
	f := newFixture(t, 5)
	order := f.createOrder(t, 1)
	_ = f.createPendingPayment(t, order)
	if _, err := f.pool.Exec(f.ctx, `UPDATE orders SET created_at=NOW()-INTERVAL '2 hours' WHERE id=$1::uuid`, order.OrderID); err != nil {
		t.Fatal(err)
	}
	svc := service.NewOrderService(repository.NewOrderRepository(f.pool))
	count, err := svc.ExpireUnpaidOrders(f.ctx, 60, 10)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("must not expire unresolved payment; expired %d", count)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 5 || reserved != 1 {
		t.Fatalf("unresolved payment reservation changed: on_hand=%d reserved=%d", onHand, reserved)
	}
}

func TestAdminReportQueries(t *testing.T) {
	f := newFixture(t, 5)
	order := f.createOrder(t, 1)
	orders, err := repository.NewReportRepository(f.pool).Orders(f.ctx)
	if err != nil {
		t.Fatalf("orders report: %v", err)
	}
	if orders.TotalOrders < 1 || orders.PendingOrders < 1 {
		t.Fatalf("report did not include test order: %+v", orders)
	}
	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC().Add(24 * time.Hour)
	sales, err := repository.NewReportRepository(f.pool).Sales(f.ctx, from, to)
	if err != nil {
		t.Fatalf("sales report: %v", err)
	}
	if sales.FromDate == "" || sales.ToDate == "" {
		t.Fatalf("sales date range missing: %+v", sales)
	}
	inventory, err := repository.NewReportRepository(f.pool).Inventory(f.ctx, 5)
	if err != nil {
		t.Fatalf("inventory report: %v", err)
	}
	if inventory.ActiveProducts < 1 {
		t.Fatalf("inventory report omitted active test product: %+v", inventory)
	}
	_ = order
}

func TestConcurrentSameKeyOrderRequestsCreateOneOrder(t *testing.T) {
	f := newFixture(t, 10)
	svc := service.NewOrderService(repository.NewOrderRepository(f.pool))
	req := domain.CreateOrderRequest{
		UserID: f.userID, IdempotencyKey: "parallel-order-key-" + f.suffix,
		Items: []domain.CreateOrderLineRequest{{ProductID: f.productID, Quantity: 1}},
	}
	const callers = 8
	start := make(chan struct{})
	type outcome struct {
		result domain.CreateOrderResult
		err    error
	}
	results := make(chan outcome, callers)
	var wg sync.WaitGroup
	wg.Add(callers)
	for i := 0; i < callers; i++ {
		go func() {
			defer wg.Done()
			<-start
			result, err := svc.CreateOrder(f.ctx, req)
			results <- outcome{result: result, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	orderID := ""
	createdCount := 0
	for outcome := range results {
		if outcome.err != nil {
			t.Fatalf("concurrent idempotent order request failed: %v", outcome.err)
		}
		if orderID == "" {
			orderID = outcome.result.OrderID
		}
		if outcome.result.OrderID != orderID {
			t.Fatalf("same key returned multiple orders: first=%s other=%s", orderID, outcome.result.OrderID)
		}
		if !outcome.result.Replayed {
			createdCount++
		}
	}
	if createdCount != 1 {
		t.Fatalf("expected exactly one non-replay result, got %d", createdCount)
	}
	var orderCount int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM orders WHERE user_id=$1::uuid AND idempotency_key=$2`, f.userID, req.IdempotencyKey).Scan(&orderCount); err != nil {
		t.Fatal(err)
	}
	if orderCount != 1 {
		t.Fatalf("expected one order row for concurrent key, got %d", orderCount)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 10 || reserved != 1 {
		t.Fatalf("concurrent retries should reserve exactly one unit, got on_hand=%d reserved=%d", onHand, reserved)
	}
}
