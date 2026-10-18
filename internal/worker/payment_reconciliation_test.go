package worker

import (
	"bytes"
	"context"
	"log"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

type fakeReporter struct{ calls atomic.Int32 }

func (f *fakeReporter) ReconciliationReport(_ context.Context, older int) (domain.PaymentReconciliationReport, error) {
	f.calls.Add(1)
	return domain.PaymentReconciliationReport{OlderThanMinutes: older, IssueCount: 1, Issues: []domain.PaymentReconciliationIssue{{IssueType: "STALE_PENDING_PAYMENT", OrderID: "test-order", Details: "synthetic test issue"}}}, nil
}

func TestPeriodicReconciliationRunsAndStops(t *testing.T) {
	reporter := &fakeReporter{}
	var output bytes.Buffer
	logger := log.New(&output, "", 0)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); RunPeriodicReconciliation(ctx, reporter, 5*time.Millisecond, 15, logger) }()
	deadline := time.Now().Add(500 * time.Millisecond)
	for reporter.calls.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if reporter.calls.Load() == 0 {
		cancel()
		<-done
		t.Fatal("expected reconciliation worker to run")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
	if !strings.Contains(output.String(), "issue_count=1") || !strings.Contains(output.String(), "STALE_PENDING_PAYMENT") {
		t.Fatalf("expected report issue in logs, got %q", output.String())
	}
}
