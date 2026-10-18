package integration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

// openOrderFlowBenchmarkFixture refuses any database other than orderflow_test.
// Benchmark-created records are tagged by a unique user/product and cleaned up.
func openOrderFlowBenchmarkFixture(b *testing.B, initialStock int) (*pgxpool.Pool, context.Context, string, string) {
	b.Helper()
	dsn := os.Getenv("ORDERFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		b.Skip("set ORDERFLOW_TEST_DATABASE_URL to the dedicated orderflow_test database")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		b.Fatalf("parse ORDERFLOW_TEST_DATABASE_URL: %v", err)
	}
	if cfg.ConnConfig.Database != integrationDatabaseName {
		b.Fatalf("refusing to benchmark database %q; expected %q", cfg.ConnConfig.Database, integrationDatabaseName)
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		cancel()
		b.Fatalf("connect benchmark database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		cancel()
		b.Fatalf("ping benchmark database: %v", err)
	}
	b.Cleanup(func() { pool.Close(); cancel() })

	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		b.Fatalf("generate benchmark suffix: %v", err)
	}
	suffix := hex.EncodeToString(raw[:])
	var userID, productID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, 'benchmark-test-only', 'CUSTOMER') RETURNING id::text
	`, fmt.Sprintf("bench-%s@example.test", suffix)).Scan(&userID); err != nil {
		b.Fatalf("create benchmark user: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO products (sku, name, description, price_paise, is_active)
		VALUES ($1, 'OrderFlow benchmark product', 'Isolated benchmark fixture', 100, TRUE)
		RETURNING id::text
	`, "BENCH-"+suffix).Scan(&productID); err != nil {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1::uuid`, userID)
		b.Fatalf("create benchmark product: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO inventory (product_id, on_hand_quantity, reserved_quantity)
		VALUES ($1::uuid, $2, 0)
	`, productID, initialStock); err != nil {
		b.Fatalf("create benchmark inventory: %v", err)
	}
	b.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM orders WHERE user_id=$1::uuid`, userID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM inventory_movements WHERE product_id=$1::uuid`, productID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM inventory WHERE product_id=$1::uuid`, productID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM products WHERE id=$1::uuid`, productID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id=$1::uuid`, userID)
	})
	return pool, ctx, userID, productID
}

func BenchmarkOrderCreation(b *testing.B) {
	pool, ctx, userID, productID := openOrderFlowBenchmarkFixture(b, 1_000_000)
	orders := service.NewOrderService(repository.NewOrderRepository(pool))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := orders.CreateOrder(ctx, domain.CreateOrderRequest{
			UserID:         userID,
			IdempotencyKey: fmt.Sprintf("bench-order-key-%d", i),
			Items:          []domain.CreateOrderLineRequest{{ProductID: productID, Quantity: 1}},
		})
		if err != nil {
			b.Fatalf("create benchmark order: %v", err)
		}
	}
	b.StopTimer()
}

func BenchmarkOrderHistoryList(b *testing.B) {
	pool, ctx, userID, productID := openOrderFlowBenchmarkFixture(b, 1_000_000)
	orders := service.NewOrderService(repository.NewOrderRepository(pool))
	// Seed a stable 100-order history outside the measured interval.
	b.StopTimer()
	for i := 0; i < 100; i++ {
		_, err := orders.CreateOrder(ctx, domain.CreateOrderRequest{
			UserID:         userID,
			IdempotencyKey: fmt.Sprintf("bench-history-key-%d", i),
			Items:          []domain.CreateOrderLineRequest{{ProductID: productID, Quantity: 1}},
		})
		if err != nil {
			b.Fatalf("seed benchmark history: %v", err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := orders.ListForUser(ctx, userID, 1, 20); err != nil {
			b.Fatalf("list benchmark order history: %v", err)
		}
	}
	b.StopTimer()
}
