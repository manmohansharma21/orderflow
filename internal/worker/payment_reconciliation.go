// Package worker contains optional background operational workers.
package worker

import (
	"context"
	"log"
	"time"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

type ReconciliationReporter interface {
	ReconciliationReport(context.Context, int) (domain.PaymentReconciliationReport, error)
}

type MockProviderOutcomeRecoverer interface {
	ReconcileKnownMockProviderOutcomes(context.Context, int, int) (int, error)
}

// RunPeriodicReconciliation recovers known terminal outcomes from the simulated
// provider ledger through the payment service's validated transition path, then
// logs remaining stale/inconsistent payment-order state. It never contacts a live
// provider and it never performs real charges or refunds.
func RunPeriodicReconciliation(ctx context.Context, reporter ReconciliationReporter, interval time.Duration, olderThanMinutes int, logger *log.Logger) {
	if interval <= 0 || reporter == nil || logger == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			if recoverer, ok := reporter.(MockProviderOutcomeRecoverer); ok {
				recovered, recoverErr := recoverer.ReconcileKnownMockProviderOutcomes(runCtx, olderThanMinutes, 100)
				if recoverErr != nil {
					logger.Printf("ERROR mock-provider outcome recovery failed: %v", recoverErr)
				} else if recovered > 0 {
					logger.Printf("mock-provider reconciliation recovered_count=%d older_than_minutes=%d", recovered, olderThanMinutes)
				}
			}
			report, err := reporter.ReconciliationReport(runCtx, olderThanMinutes)
			cancel()
			if err != nil {
				logger.Printf("ERROR payment reconciliation failed: %v", err)
				continue
			}
			logger.Printf("payment reconciliation completed issue_count=%d older_than_minutes=%d", report.IssueCount, report.OlderThanMinutes)
			for _, issue := range report.Issues {
				logger.Printf("payment reconciliation issue type=%s payment_id=%s order_id=%s payment_status=%s order_status=%s details=%q", issue.IssueType, issue.PaymentID, issue.OrderID, issue.PaymentStatus, issue.OrderStatus, issue.Details)
			}
		}
	}
}
