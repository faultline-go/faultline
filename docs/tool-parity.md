# Tool Parity

Faultline's OSS scanner intentionally delegates language-specific checks where
the Go toolchain already has a stable, trusted implementation. This keeps the
native scanner focused on structural risk, ownership, coverage, dependency
shape, suppressions, and export contracts.

## CI Parity Gates

| Gate | Command | Source of truth |
|---|---|---|
| Formatting | `gofmt -l` through `make fmt` and CI formatting checks | Go formatter |
| Module graph hygiene | `go mod tidy` plus `git diff --exit-code -- go.mod go.sum` | Go modules |
| Tests | `go test ./...` | Go test runner |
| Vet diagnostics | `go vet ./...` | Go vet |
| Lint parity | `golangci-lint run` | GolangCI-Lint configuration |
| Dead code parity | `bash scripts/deadcode-check.sh` | `golang.org/x/tools/cmd/deadcode` with the committed baseline |
| Release build parity | `goreleaser release --snapshot --clean` | GoReleaser configuration |

`make quality` runs formatting, module tidy checking, tests, GolangCI-Lint, and
dead-code analysis. CI also runs `go vet`, release helper checks, CGO-disabled
builds, Docker builds, config validation, and a Faultline self-scan.

## Faultline-Native Signals

Faultline owns these scanner signals directly:

- package risk score calculation and component score normalization
- coverage profile ingestion and package coverage gaps
- git churn and recent author diversity
- CODEOWNERS and configured module ownership evidence
- dependency structure findings from local `go.mod` and loaded imports
- architecture boundary findings from `faultline.yaml`
- suppression validation and suppression audit metadata
- snapshot export shape for `faultline.snapshot.v1`

Faultline does not replace `gofmt`, `go vet`, GolangCI-Lint, `deadcode`,
GoReleaser, SBOM tooling, Cosign, or GitHub artifact attestations. Those tools
remain the source of truth for their domains.

## Dogfood Contract

The public repository must be clean enough to scan itself before v1.0:

```sh
go test ./... -coverprofile=coverage.out
go run ./cmd/faultline config validate --config .faultline.yaml --strict
go run ./cmd/faultline scan ./... \
  --config .faultline.yaml \
  --strict-config \
  --coverage coverage.out \
  --format snapshot \
  --out faultline-snapshot.json \
  --no-history
jq -e '.schema_version == "faultline.snapshot.v1" and (.packages | type == "array")' faultline-snapshot.json
```

The self-scan gate validates that the committed config schema is current, that
coverage can be consumed by the scanner, and that the emitted snapshot remains a
machine-readable `faultline.snapshot.v1` document.
