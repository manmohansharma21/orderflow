# OrderFlow Finalization Guide

This repository is the source of truth. Incremental ZIPs shared during development are overlays only; they are not the upload artifact for the Scaler project ZIP field.

## Current baseline

- The latest verified baseline commit before this final-review patch is `d891056`.
- Its GitHub Actions run succeeded: [Go CI run](https://github.com/manmohansharma21/orderflow/actions/runs/39156676397).
- Product-read cache and order/payment reliability features are committed.
- The measured local product-list API benchmark showed Redis slower than direct PostgreSQL for the tested 300-request workload. Redis is therefore optional and disabled by default until a representative workload proves otherwise.

## Apply this final-review patch

From the repository root:

```bash
unzip -o ~/Downloads/ORDERFLOW_FINAL_REVIEW_PATCH.zip -d .
chmod +x scripts/apply-final-review-patch.sh
./scripts/apply-final-review-patch.sh
```

The patch adds separate simulated-provider state, missed-callback recovery through the shared payment transition path, request IDs and HTTP histogram metrics, a Postman recovery demo, updated OpenAPI/docs/diagrams and the actual benchmark output. It also removes committed report draft binaries from source control because the final PDF must be uploaded separately.

Do not describe these additions as complete until the verification checklist passes and the new commit has green CI.

## Submission fields

The Scaler page expects a public GitHub URL, a full project ZIP under its size limit, and a separate report PDF. After committing/pushing all final code and evidence, download the complete source archive from:

`https://github.com/manmohansharma21/orderflow/archive/refs/heads/main.zip`

Do not upload this incremental patch ZIP as the project ZIP. See `docs/FINAL_SUBMISSION_GUIDE.md`.
