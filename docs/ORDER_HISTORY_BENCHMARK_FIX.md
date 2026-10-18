# Order-history benchmark fix

The previous `BenchmarkOrderHistoryList` called `b.StopTimer()` before seeding data and then `b.ResetTimer()` without restarting the timer. In Go's benchmark API, `ResetTimer` does not resume a stopped timer, so the benchmark harness could keep increasing `b.N` without measuring time and appear to hang.

The fixed benchmark explicitly calls `b.StartTimer()` after setup/reset. The script uses `-benchtime=200ms -count=1` to keep this evidence run bounded. Results are service/repository-path measurements, not HTTP p95.

The previously interrupted result file `docs/evidence/performance/order-history-benchmark-20261018T201742Z.txt` is incomplete and must not be presented as successful benchmark evidence. Run `scripts/clean-incomplete-benchmark-artifact.sh` to remove that known partial log and clean whitespace previously flagged by `git diff --check`.
