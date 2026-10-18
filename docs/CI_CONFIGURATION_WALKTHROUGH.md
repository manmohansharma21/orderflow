# CI configuration walkthrough — learn at final review

CI means **Continuous Integration**: each push/pull request triggers an automated quality gate so a change is tested in a repeatable environment rather than relying only on a developer's laptop.

Open `.github/workflows/ci.yml` from top to bottom:

1. **Triggers:** `push` and `pull_request` targeting `main` mean the workflow runs on new main-branch commits and proposed changes.
2. **Permissions:** `contents: read` gives the job only the repository read access needed for checkout.
3. **Runner and service containers:** `ubuntu-latest` runs the job; service containers create disposable PostgreSQL and Redis. Their health checks ensure dependencies are ready before tests.
4. **Configuration:** workflow environment variables point database tests at the database named `orderflow_test`, and Redis tests at the temporary service. Never copy CI test credentials into production.
5. **Module consistency:** `go mod tidy` followed by a diff check ensures dependencies declared in `go.mod`/`go.sum` agree with the source.
6. **Migrations:** the numbered `migrations/*.up.sql` files run in ascending filename order against the disposable CI DB with stop-on-error enabled.
7. **Tests:** `go test -count=1 ./...` runs tests without trusting cached test results; integration race checks, `go vet`, and whitespace checks add separate gates.
8. **Image build:** the `Dockerfile` is built in CI to catch packaging/compiler issues.
9. **Runtime smoke test:** the built container starts and the workflow waits for `/health` and checks `/ready`. This validates a runnable image, not a live AWS deployment.
10. **Result/evidence:** for submissions, capture the completed green run and the commit SHA that was tested. If source changes after that commit, wait for the new run.

Latest reviewed green run: https://github.com/manmohansharma21/orderflow/actions/runs/39160539244 (commit `1515cb0`). If a later source commit becomes `main`, this run remains evidence for `1515cb0`, not for that future commit.
