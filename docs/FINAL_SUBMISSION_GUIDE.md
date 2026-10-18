# Scaler/Woolf capstone submission guide

The screenshot of the submission page shows three required deliverable fields: (1) GitHub Repo Link, (2) Upload Project as a ZIP up to 50 MB with proper source structure and README, and (3) Upload Project Report as a PDF up to 50 MB. It also shows two declaration checkboxes. Review and tick those yourself only when each statement is true.

## What to upload

- **GitHub Repo Link:** `https://github.com/manmohansharma21/orderflow`
- **Upload Project:** the full source archive from the latest pushed `main`, not the incremental ZIP patches used during development.
- **Upload Project Report:** the final personalized, evidence-complete PDF exported from the Word report. The portal's report field expects PDF; the DOCX is for editing/review.

Full source archive link: <https://github.com/manmohansharma21/orderflow/archive/refs/heads/main.zip>

## Required order

1. Complete code/docs edits and actual screenshots/benchmark output.
2. `git status --short`, `git diff --check`, stage only intentional source/docs/evidence. Never stage `.env`, tokens, secrets, DB backups or local personal files.
3. Commit/push the last revision. Confirm the newest Actions run is green and its SHA matches `git rev-parse --short HEAD`.
4. Only then download the current `main.zip` from GitHub. This archive includes whatever is in remote `main` at download time.
5. Verify source ZIP structure, README, migrations, tests, diagrams, OpenAPI and Postman collection; ensure no `.env`, passwords, database dumps, access tokens or report drafts are included.
6. Complete the Word report: insert evidence PNGs, refresh figure/table list and page numbers, replace only the known personal date placeholders with actual dates, validate citations/AI-tool acknowledgement, and proofread every page.
7. Export to PDF and render/check all pages. Keep the DOCX separately for editing.
8. Upload the repo URL, full source ZIP and final PDF in their respective fields. Read the declarations and tick them only if accurate.

## Report personalization safeguards

The editable review draft uses the candidate name Manmohan Sharma, the supplied mentee email, and Naman Bhalla as supervisor from the provided template. The actual project module start/end dates and actual submission date are intentionally placeholders because they cannot be inferred safely. Candidate signature must be supplied by the candidate. Do not complete the supervisor certification on the supervisor's behalf. Disclose AI-assisted drafting/review and any external sources according to the policy applicable to the programme.

## Scope claims to avoid

- Do not claim a live AWS deployment unless real resources were deployed, smoke-tested and evidence collected.
- Do not claim Redis improved performance: the provided measured local run showed the warm cache slower.
- Do not claim a real Stripe integration: the implementation uses a mock provider.
- Do not invent screenshots, benchmark results, reviewer approval or sign-off.
