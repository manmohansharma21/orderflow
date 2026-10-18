package integration

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

func TestOrderCreationConcurrentRequestsNeverOversellStock(t *testing.T) {
	const (
		initialStock = 5
		requestCount = 10
	)
	f := newFixture(t, initialStock)
	orderService := service.NewOrderService(repository.NewOrderRepository(f.pool))

	start := make(chan struct{})
	errorsCh := make(chan error, requestCount)
	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := orderService.CreateOrder(f.ctx, domain.CreateOrderRequest{
				UserID: f.userID, ProductID: f.productID, Quantity: 1,
			})
			errorsCh <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errorsCh)

	successes := 0
	insufficient := 0
	for err := range errorsCh {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, service.ErrInsufficientStock):
			insufficient++
		default:
			t.Errorf("unexpected concurrent order error: %v", err)
		}
	}
	if successes != initialStock || insufficient != requestCount-initialStock {
		t.Fatalf("expected %d successes and %d insufficient-stock responses; got %d successes and %d insufficient-stock responses", initialStock, requestCount-initialStock, successes, insufficient)
	}

	onHand, reserved := f.inventory(t)
	if onHand != initialStock || reserved != initialStock {
		t.Fatalf("concurrent reservations oversold or lost stock: on_hand=%d reserved=%d", onHand, reserved)
	}
	var orderCount, movementCount int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1::uuid`, f.userID).Scan(&orderCount); err != nil {
		t.Fatalf("count concurrently created orders: %v", err)
	}
	if err := f.pool.QueryRow(f.ctx, `
		SELECT COUNT(*) FROM inventory_movements
		WHERE product_id = $1::uuid AND movement_type = 'ORDER_RESERVED'
	`, f.productID).Scan(&movementCount); err != nil {
		t.Fatalf("count concurrent reservation movements: %v", err)
	}
	if orderCount != initialStock || movementCount != initialStock {
		t.Fatalf("expected %d committed orders and reservation movements; got orders=%d movements=%d", initialStock, orderCount, movementCount)
	}
}

func TestOrderCreationRollsBackReservationWhenOrderInsertCannotFindUser(t *testing.T) {
	f := newFixture(t, 4)
	orderService := service.NewOrderService(repository.NewOrderRepository(f.pool))
	missingUserID := "00000000-0000-0000-0000-000000000000"

	_, err := orderService.CreateOrder(f.ctx, domain.CreateOrderRequest{
		UserID: missingUserID, ProductID: f.productID, Quantity: 2,
	})
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound after the reservation attempt, got %v", err)
	}

	onHand, reserved := f.inventory(t)
	if onHand != 4 || reserved != 0 {
		t.Fatalf("transaction did not roll back reservation: on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, "00000000-0000-0000-0000-000000000000", "ORDER_RESERVED"); got != 0 {
		t.Fatalf("rollback left inventory movement rows referencing the missing order: %d", got)
	}
	var orderCount, movementCount int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1::uuid`, missingUserID).Scan(&orderCount); err != nil {
		t.Fatalf("count orders for missing user: %v", err)
	}
	if err := f.pool.QueryRow(f.ctx, `
		SELECT COUNT(*) FROM inventory_movements
		WHERE product_id = $1::uuid AND movement_type = 'ORDER_RESERVED'
	`, f.productID).Scan(&movementCount); err != nil {
		t.Fatalf("count reservation movements after rollback: %v", err)
	}
	if orderCount != 0 || movementCount != 0 {
		t.Fatalf("rollback left partial records: orders=%d movements=%d", orderCount, movementCount)
	}
}

func TestConcurrentPaymentInitiationWithSameKeyCreatesOnePayment(t *testing.T) {
	const requestCount = 10
	f := newFixture(t, 3)
	order := f.createOrder(t, 1)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	key := "concurrent-key-" + f.suffix

	type result struct {
		id  string
		err error
	}
	start := make(chan struct{})
	results := make(chan result, requestCount)
	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			payment, err := payments.CreateForOrder(f.ctx, f.userID, order.OrderID, key)
			results <- result{id: payment.ID, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	paymentID := ""
	for result := range results {
		if result.err != nil {
			t.Errorf("concurrent payment initiation failed: %v", result.err)
			continue
		}
		if paymentID == "" {
			paymentID = result.id
		}
		if result.id != paymentID {
			t.Errorf("same idempotency key returned different payment IDs: %s and %s", paymentID, result.id)
		}
	}
	if paymentID == "" {
		t.Fatal("no concurrent payment initiation returned a payment ID")
	}

	var count int
	if err := f.pool.QueryRow(f.ctx, `SELECT COUNT(*) FROM payments WHERE order_id = $1::uuid`, order.OrderID).Scan(&count); err != nil {
		t.Fatalf("count payment records after concurrent initiation: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one payment record, got %d", count)
	}
}

func TestConcurrentFailedPaymentCallbackReleasesStockOnlyOnce(t *testing.T) {
	const requestCount = 10
	f := newFixture(t, 6)
	order := f.createOrder(t, 2)
	paymentID := f.createPendingPayment(t, order)
	payments := service.NewPaymentService(repository.NewPaymentRepository(f.pool))
	eventID := fmt.Sprintf("concurrent-failure-%s", f.suffix)

	type result struct {
		status string
		err    error
	}
	start := make(chan struct{})
	results := make(chan result, requestCount)
	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			payment, err := payments.ProcessMockCallback(f.ctx, paymentID, "FAILED", eventID)
			results <- result{status: payment.Status, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	for result := range results {
		if result.err != nil {
			t.Errorf("concurrent replay of failed callback returned an error: %v", result.err)
			continue
		}
		if result.status != "FAILED" {
			t.Errorf("expected callback result status FAILED, got %q", result.status)
		}
	}
	if got := f.orderStatus(t, order.OrderID); got != "FAILED" {
		t.Fatalf("expected failed order status, got %s", got)
	}
	onHand, reserved := f.inventory(t)
	if onHand != 6 || reserved != 0 {
		t.Fatalf("concurrent callback replay changed inventory unexpectedly: on_hand=%d reserved=%d", onHand, reserved)
	}
	if got := f.movementCount(t, order.OrderID, "ORDER_RELEASED"); got != 1 {
		t.Fatalf("expected exactly one ORDER_RELEASED movement after concurrent callback replay, got %d", got)
	}
}
