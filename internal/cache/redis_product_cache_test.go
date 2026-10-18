package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// This test only touches Redis DB 14 on a dedicated disposable Redis instance.
func TestRedisProductCacheTTLAndVersion(t *testing.T) {
	addr := os.Getenv("ORDERFLOW_REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("set ORDERFLOW_REDIS_TEST_ADDR to run Redis adapter integration test")
	}
	c := NewRedisProductCache(addr, "", 14)
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		t.Fatalf("ping dedicated test Redis: %v", err)
	}
	key := fmt.Sprintf("orderflow:test:%d", time.Now().UnixNano())
	if err := c.Set(ctx, key, []byte("cached"), 50*time.Millisecond); err != nil {
		t.Fatalf("set cache value: %v", err)
	}
	got, err := c.Get(ctx, key)
	if err != nil || string(got) != "cached" {
		t.Fatalf("unexpected cached value %q, err=%v", got, err)
	}
	before, err := c.Version(ctx)
	if err != nil {
		t.Fatalf("read cache version: %v", err)
	}
	if err := c.InvalidateProducts(ctx); err != nil {
		t.Fatalf("invalidate cache: %v", err)
	}
	after, err := c.Version(ctx)
	if err != nil {
		t.Fatalf("read invalidated version: %v", err)
	}
	if after != before+1 {
		t.Fatalf("expected generation %d, got %d", before+1, after)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := c.Get(ctx, key); err == nil {
		t.Fatal("expected cached value to expire by TTL")
	}
}
