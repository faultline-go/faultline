# Tool Parity

Faultline reports structural governance evidence. It is not a replacement for Go's native static-analysis tools.

## Native Go Gates

| Gate | Tool | Faultline Role |
| --- | --- | --- |
| Formatting | `gofmt` | CI parity gate only |
| Unit tests | `go test ./...` | Coverage input and release quality gate |
| Vet checks | `go vet ./...` | CI parity gate only |
| Dead code | `golang.org/x/tools/cmd/deadcode` | CI parity gate only |
| Linting | `golangci-lint` | CI parity gate only |

## Faultline-Specific Evidence

| Signal | Source | Output |
| --- | --- | --- |
| Ownership | CODEOWNERS, config, git authorship | ownership findings and owner confidence |
| Churn | local git history | churn score and findings |
| Coverage | Go coverage profile | coverage gap score and findings |
| Package structure | `go/packages`, imports, module discovery | dependency centrality and boundary findings |
| Suppressions | `faultline.yaml` | local waiver audit and Enterprise metadata |
| Snapshot | `pkg/export` | `faultline.snapshot.v1` source-free contract |
