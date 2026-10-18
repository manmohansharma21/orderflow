package integration

import (
	"os"
	"testing"
	"time"

	"github.com/manmohansharma21/orderflow/internal/cache"
	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

// Uses dedicated PostgreSQL integration DB orderflow_test and Redis DB 15 only.
func TestProductReadCacheInvalidatedAfterMutation(t *testing.T) {
	addr := os.Getenv("ORDERFLOW_REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("set ORDERFLOW_REDIS_TEST_ADDR to run Redis+PostgreSQL product-cache test")
	}
	f := newFixture(t, 4)
	rc := cache.NewRedisProductCache(addr, "", 15)
	t.Cleanup(func() { _ = rc.Close() })
	if err := rc.Ping(f.ctx); err != nil {
		t.Fatalf("ping test Redis: %v", err)
	}
	svc := service.NewProductServiceWithCache(repository.NewProductRepository(f.pool), rc, time.Minute)
	list, err := svc.List(f.ctx, "IT-"+f.suffix, 1, 20)
	if err != nil || len(list.Items) != 1 {
		t.Fatalf("warm product list cache: count=%d err=%v", len(list.Items), err)
	}
	old := list.Items[0].Name
	newName := "Updated integration product " + f.suffix
	if _, err := svc.Update(f.ctx, f.productID, domain.UpdateProductInput{Name: &newName}); err != nil {
		t.Fatalf("update product: %v", err)
	}
	updatedList, err := svc.List(f.ctx, "IT-"+f.suffix, 1, 20)
	if err != nil {
		t.Fatalf("read product list after mutation: %v", err)
	}
	if len(updatedList.Items) != 1 || updatedList.Items[0].Name != newName {
		t.Fatalf("stale product-list cache after mutation (old=%q): %+v", old, updatedList.Items)
	}
	updatedItem, err := svc.GetByID(f.ctx, f.productID)
	if err != nil || updatedItem.Name != newName {
		t.Fatalf("stale product-detail cache after mutation: item=%+v err=%v", updatedItem, err)
	}
}
