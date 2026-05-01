# Faultline Demo

This demo uses the bundled fixture in `testdata/simple-go-module`. It is small
enough to inspect quickly, but it exercises the same report pipeline used for
real repositories.

## Run It

From the repository root:

```sh
go build -o bin/faultline ./cmd/faultline
cd testdata/simple-go-module
../../bin/faultline scan ./... --format html --out faultline-report.html --no-history
../../bin/faultline scan ./... --format json --out faultline-report.json --no-history
```

Open `faultline-report.html` in a browser.

## Example Output

Generated examples are committed for quick review:

- [HTML report](../examples/reports/simple-go-module.html)
- [JSON report](../examples/reports/simple-go-module.json)

The sample intentionally runs without a coverage profile, so the report includes
an informational warning that coverage is unknown. Faultline treats this as a
signal to make missing coverage visible, not as 0% coverage.

## What To Look For

- Scan metadata: version, repository path, config hash, package patterns.
- Summary cards: package count, warnings, generated file percentage.
- Package risk table: normalized risk score and component evidence.
- Findings: category, severity, recommendation, confidence, and evidence.
- Evidence appendix: the raw facts used to explain package scores.

## Real Repo Command

For a real Go repository:

```sh
go test ./... -coverprofile=coverage.out
faultline scan ./... --coverage coverage.out --format html --out faultline-report.html
faultline scan ./... --coverage coverage.out --format sarif --out faultline.sarif
```

Faultline does not upload source code or call remote services during default
scans. Optional integrations such as SARIF upload are controlled by your CI
workflow, not by the scanner.
