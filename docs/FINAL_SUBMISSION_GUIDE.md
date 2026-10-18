# Scaler capstone submission guide

The submission page in the supplied screenshot has three deliverable inputs: a public GitHub repository link, a project ZIP (maximum 50 MB) with correct repository structure and README, and a separate report PDF written using the Woolf report template. The page also has two declarations/checkboxes. Review and tick those yourself only when they are true for your work.

## What to enter

- **GitHub Repo Link:** `https://github.com/manmohansharma21/orderflow`
- **Upload Project:** use a full source archive of the *latest pushed `main` branch*, not an incremental overlay ZIP.
- **Upload Project Report:** use the final personalized PDF exported from the report DOCX after the evidence session. Do not submit a `DRAFT` file.

Direct full-source archive (this is a link to GitHub's downloadable archive, not an incremental code patch):

`https://github.com/manmohansharma21/orderflow/archive/refs/heads/main.zip`

The archive automatically tracks the current remote branch. Therefore, commit/push the final source and actual benchmark output first, wait for CI to go green for that exact SHA, and only then download the source archive for the portal.

## Before uploading

1. Confirm the README, HLD/LLD sources, migration files, tests and OpenAPI/Postman collection are present in GitHub `main`.
2. Confirm the newest GitHub Actions run is green and its SHA matches `git rev-parse --short HEAD`.
3. Confirm there is no `.env`, token, password, database dump, personal access token, or generated secret in the source archive.
4. Remove draft report binaries from the source repository; the PDF belongs in the separate report upload field.
5. Fill the actual registered student email, actual submission date, and project module dates in the Word report. Review the declaration, references and AI/tool acknowledgement for accuracy. Do not fill the supervisor certification/signature on behalf of the supervisor.
6. Complete the end-stage screenshot checklist with real browser/Postman/CI evidence and update the report page/figure references.
7. Verify the final PDF is readable and meets the template formatting/length requirements before uploading.

## Do not claim

- Do not claim a live AWS deployment unless it was provisioned and smoke-tested.
- Do not claim Redis improved latency: the measured local result above was slower for the current dataset.
- Do not claim a real Stripe/payment integration. The project scope uses a simulated payment provider.
- Do not fabricate endpoint screenshots, CI runs or metrics. Use actual captured output.
