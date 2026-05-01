# Product Hunt Draft

## Name

```text
Faultline
```

## Tagline

```text
Local-first structural risk reports for Go codebases
```

## Description

```text
Faultline is an open-source CLI that scans Go repositories and produces
explainable package-level risk reports from code structure, git churn, coverage,
ownership, dependency metadata, and architecture policy rules.

It runs locally, does not upload source code, and emits HTML, JSON, and SARIF
for GitHub code scanning. It also supports baselines, local history, rule packs,
suppressions, PR review markdown, and multi-module Go repositories.
```

## Maker Comment

```text
Faultline started from a practical problem: repo-level quality checks often miss
the packages that are risky because they are central, changing fast, ownerless,
under-tested, or crossing architecture boundaries.

The CLI is local-first and open source. You can run:

go install github.com/faultline-go/faultline/cmd/faultline@latest
faultline scan ./... --format html --out faultline-report.html

No source code leaves your machine. The current outputs are HTML, JSON, and
SARIF, with local baselines/history for risk trends.

I’m looking for feedback from Go teams, platform engineers, and maintainers of
larger repositories: which findings are useful, which are noisy, and what would
make this fit better into CI review workflows?
```

## Assets To Prepare

- Social preview image from `docs/assets/social-preview.svg`.
- Screenshot or GIF of `examples/reports/simple-go-module.html`.
- Link to repository README.
- Link to `docs/demo.md`.
