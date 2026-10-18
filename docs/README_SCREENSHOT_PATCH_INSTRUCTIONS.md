# Applying this README / report-source / screenshot guide patch

This is an **incremental documentation overlay**. Extract it into the repository root, review the changes, then commit/push. It is not the full source ZIP for Scaler.

```bash
unzip -o ~/Downloads/ORDERFLOW_README_SCREENSHOT_GUIDE_PATCH.zip -d .
git status --short
git diff --check
git diff -- README.md docs/PROJECT_REPORT.md docs/SCREENSHOT_CHECKLIST.md docs/evidence/README.md docs/FINALIZATION_STATUS.md docs/FINAL_SUBMISSION_GUIDE.md docs/CI_CONFIGURATION_WALKTHROUGH.md
```

Review the content and the actual benchmark values. Do not add the Word report review draft to the repository; upload the report PDF separately via the Scaler submission page. Then stage the intended docs, inspect `git diff --cached --stat`, commit and push. Wait for green CI on the resulting commit SHA before downloading the final GitHub source archive.

The documentation deliberately marks screenshots, the complete order-history benchmark, report-PDF finalization, live AWS and the prior Lambda review as incomplete/conditional. Do not remove those caveats unless real evidence supports doing so.
