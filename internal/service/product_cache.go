package service

import (
	"context"
	"time"
)

// ProductCache stores JSON-encoded product read models. Implementations are
// optional: cache errors must never make the product API unavailable.
type ProductCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Version(ctx context.Context) (int64, error)
	InvalidateProducts(ctx context.Context) error
}
