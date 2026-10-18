package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manmohansharma21/orderflow/internal/cache"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

// BenchmarkProductListDatabaseVsRedis compares the same product-list query with
// cache disabled and with a warmed Redis cache. It intentionally refuses to use
// any PostgreSQL DB other than orderflow_test and records no fabricated values.
func BenchmarkProductListDatabaseVsRedis(b *testing.B) {
	dsn := os.Getenv("ORDERFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		b.Skip("set ORDERFLOW_TEST_DATABASE_URL to dedicated orderflow_test")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		b.Fatal(err)
	}
	if cfg.ConnConfig.Database != integrationDatabaseName {
		b.Fatalf("refusing to benchmark database %q; expected %q", cfg.ConnConfig.Database, integrationDatabaseName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		b.Fatal(err)
	}
	sku := fmt.Sprintf("BENCH-%d", time.Now().UnixNano())
	var productID string
	if err := pool.QueryRow(ctx, `INSERT INTO products(sku,name,price_paise,is_active) VALUES($1,'Benchmark Product',49900,TRUE) RETURNING id::text`, sku).Scan(&productID); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM products WHERE id=$1::uuid`, productID) })

	repo := repository.NewProductRepository(pool)
	dbService := service.NewProductService(repo)
	addr := os.Getenv("ORDERFLOW_REDIS_TEST_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rc := cache.NewRedisProductCache(addr, "", 15)
	defer rc.Close()
	pingCtx, pingCancel := context.WithTimeout(ctx, time.Second)
	err = rc.Ping(pingCtx)
	pingCancel()
	if err != nil {
		b.Skipf("Redis benchmark skipped; start the Redis test service: %v", err)
	}
	cacheService := service.NewProductServiceWithCache(repo, rc, time.Minute)
	_ = rc.InvalidateProducts(ctx)
	if _, err := cacheService.List(ctx, "", 1, 20); err != nil {
		b.Fatal(err)
	} // warm cache

	b.Run("PostgreSQL_no_cache", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := dbService.List(ctx, "", 1, 20); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Redis_warm_cache", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := cacheService.List(ctx, "", 1, 20); err != nil {
				b.Fatal(err)
			}
		}
	})
}
