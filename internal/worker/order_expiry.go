package worker

import (
	"context"
	"log"
	"time"

	"github.com/manmohansharma21/orderflow/internal/service"
)

// RunPeriodicOrderExpiry expires old PENDING orders that have no PENDING or SUCCESS payment.
func RunPeriodicOrderExpiry(ctx context.Context, orders *service.OrderService, interval time.Duration, olderThanMinutes, batchSize int, logger *log.Logger) {
	if interval <= 0 || orders == nil || logger == nil {
		return
	}
	run := func() {
		runCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		count, err := orders.ExpireUnpaidOrders(runCtx, olderThanMinutes, batchSize)
		if err != nil {
			logger.Printf("ERROR unpaid-order expiry failed: %v", err)
			return
		}
		if count > 0 {
			logger.Printf("unpaid-order expiry completed expired_count=%d older_than_minutes=%d", count, olderThanMinutes)
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
