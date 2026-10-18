# Historical screenshot archive

These images are authentic earlier local-development captures supplied during project work. They are preserved for context, not all should appear in the final academic report.

- `01-docker-compose-ps-redacted-crop.png` is a crop of the Docker Compose portion of an earlier VS Code screenshot. The original image is deliberately excluded because it visibly exposed a local `DATABASE_URL` containing a password.
- `02-early-migration-only-not-final.png` shows only the initial migration (`000001`) and three tables at that early stage. It is historical evidence only and does not represent the final schema after migrations `000001` through `000008`.
- `03-readiness-body-only-duplicate.png` is an earlier browser response without the HTTP status line; the stronger combined `browser/01-health-readiness-http-200.png` capture is used in the report.

The original secret-bearing screenshot is not included. Never commit `.env`, database credentials, tokens, webhook secrets, or screenshots that expose them. If the displayed password was reused outside the disposable local environment, rotate it.
