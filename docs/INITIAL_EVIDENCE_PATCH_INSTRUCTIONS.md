# OrderFlow initial screenshot evidence patch

Extract this ZIP at the repository root. It preserves the actual screenshots provided during the local test session in `docs/evidence/`.

## Apply

```bash
unzip -o ~/Downloads/ORDERFLOW_INITIAL_EVIDENCE_PATCH.zip -d .
python3 scripts/apply-initial-evidence-update.py
git status --short
git diff --check
```

The ZIP adds the actual screenshots, updates the evidence capture guide, and provides a script to correct outdated status/benchmark wording in the Markdown documentation. It is an evidence/documentation overlay, not the full project ZIP for the Scaler portal.

The four received screenshots cover local health, readiness, `/metrics`, and an initial development server log. They do not show Docker image-build output, full Postman business flows, a final green CI screen, or AWS deployment.
