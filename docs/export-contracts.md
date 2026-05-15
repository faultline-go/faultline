# Faultline Export Contracts

Faultline's commercial integration boundary is metadata-only. The public CLI can
produce full local reports, and downstream systems can convert those reports into
stable snapshots using `pkg/export`.

## Snapshot Schema

Current schema version:

```text
faultline.snapshot.v1
```

The snapshot is designed for:

- paid cloud ingestion
- self-hosted enterprise control planes
- portfolio analytics
- policy governance workflows
- future signed upload bundles

It is not designed to preserve every local report detail. Local HTML, JSON, and
SARIF remain the authoritative developer-facing artifacts.

## Privacy Boundary

Snapshots include:

- repo fingerprint and display name
- config hash and rule-pack hashes
- package import path and module identity
- risk scores and component scores
- finding identity, category, severity, and metadata evidence
- owner, churn, coverage percentage, trend, and suppression metadata
- dependency module metadata and structural risk flags

Snapshots omit:

- raw source code
- full file contents
- local absolute source paths
- environment variables
- credentials
- arbitrary command output

## Compatibility

Commercial services should version ingestion by `schema_version`, not by the
internal Go package layout. The enterprise backend should support at least the
latest two minor versions of the snapshot schema after the first paid release.

## v1.0 Compatibility Promise

For the v1.0 line, `faultline.snapshot.v1` is the stable metadata-only
integration contract. Patch and minor releases may add optional fields, add new
finding IDs, or populate previously empty optional arrays, but they must not
rename existing JSON fields, change field meanings, change required top-level
types, or add source-code-bearing fields to the snapshot.

Any incompatible export change requires a new `schema_version`. Consumers should
ignore unknown fields and should reject unknown schema versions unless they have
explicitly added support for that version.

## Example Usage

```go
data, err := os.ReadFile("faultline-report.json")
if err != nil {
    return err
}

snapshot, err := export.FromReportJSON(data)
if err != nil {
    return err
}

payload, err := export.MarshalJSON(snapshot)
if err != nil {
    return err
}
```

The preferred operational flow is:

```sh
faultline scan ./... --format snapshot --out faultline-snapshot.json
```
