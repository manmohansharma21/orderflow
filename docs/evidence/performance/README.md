# Product-list benchmark evidence

The PNG `benchmark-results-20261018T230232Z.png` is the user-captured VS Code screenshot of the completed 300-request-per-variant API benchmark. The adjacent `.txt` file records the same displayed values in searchable text.

This is a second run and differs from the earlier committed run. In this run the warm Redis case was faster (P50 2.429 ms vs 2.829 ms, P95 4.140 ms vs 5.203 ms; throughput 3297.18 vs 2190.99 req/s). In an earlier run the warm Redis case was slower. Because the local results are inconsistent and the dataset is small, report both runs and do not claim a stable general improvement. Redis remains optional and disabled by default unless the configuration is intentionally changed for a test.
