# Show HN Draft

## Title

```text
Show HN: Faultline – local structural risk reports for Go codebases
```

## First Comment

```text
I built Faultline to answer a question I kept seeing in Go repos:
which packages are structurally risky before they become incidents?

It scans locally and combines package structure, git churn, coverage,
CODEOWNERS, dependency centrality, architecture boundary rules, and Go module
dependency metadata into explainable package-level reports.

No source code leaves your machine. Outputs: HTML, JSON, and SARIF.
It also supports baselines, local SQLite history, PR markdown reviews,
rule packs, suppressions, multi-module repos, and GitHub code scanning via SARIF.

Example:

go install github.com/faultline-go/faultline/cmd/faultline@latest
faultline scan ./... --format html --out faultline-report.html

I’d especially like feedback on:
- whether the risk model is explainable enough
- where the findings feel noisy
- what Go monorepo workflows are missing
- whether the local-first/privacy posture is clear enough
```

## Response Notes

- Be explicit that Faultline is not a vulnerability scanner.
- Explain that unknown coverage is visible but not treated as 0%.
- Mention that local scanning, SARIF, baselines, history, and PR markdown remain
  OSS.
- If someone challenges scoring weights, agree that calibration needs real repo
  data and point to configurable `scoring` thresholds.
- If someone asks about cloud, say the intended paid boundary is multi-repo
  governance and workflow automation, not crippling local scanning.
