# How to understand OrderFlow CI

CI means Continuous Integration: each push to `main` and pull request runs automated checks on a clean GitHub-hosted runner. It catches dependency drift, migration errors, compile/test failures and container startup failures before a change is treated as ready.

Read `.github/workflows/ci.yml` from top to bottom:

1. **Triggers**: pushes and pull requests to `main` start the workflow.
2. **Least privilege**: `contents: read` permits checkout but not write access.
3. **Service containers**: PostgreSQL 16 and Redis 7 are started as test dependencies; health checks wait for readiness.
4. **Environment variables**: tests target `orderflow_test`; Redis tests use a separate test DB index/address. These are CI-only credentials, not production secrets.
5. **Checkout and Go setup**: the runner checks out the commit and selects the Go toolchain pinned in `go.mod`.
6. **Module consistency**: `go mod tidy` followed by `git diff --exit-code -- go.mod go.sum` makes dependency metadata drift fail CI.
7. **Migrations**: the workflow applies every `migrations/*.up.sql` file in lexical order to a disposable database. A migration error stops the job.
8. **Test suites**: `go test -count=1 ./...` runs unit and database/Redis integration tests without using cached results.
9. **Race detector**: integration tests run under `go test -race` to catch Go memory races, while database locking assertions check inventory concurrency behavior.
10. **Static checks**: `go vet ./...` and `git diff --check` catch suspicious Go constructs and whitespace errors.
11. **Image build and runtime smoke test**: Docker builds the API; a container must respond successfully to `/health` and `/ready` before CI passes.

A green run is evidence for that exact commit SHA—not every future commit, AWS deployment, or business rule not covered by tests. After a push, open Actions and confirm the newest run's commit SHA matches `git log -1 --oneline`.

## Related pending learning tasks

- Review the earlier Lambda deployment submission and write down its architecture, deployment configuration and lessons before claiming it is complete.
- After submission, walk through `cmd/api`, config, HTTP DTOs/handlers, services, repositories, migrations, Redis adapter, workers and tests. The report should distinguish what is implemented from what is only planned.
