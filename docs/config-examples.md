# Faultline Config Examples

These examples are sanitized starting points for common Go repository shapes.
Keep repository-specific policy in `faultline.yaml` and shared defaults in local
rule packs. Faultline does not fetch config or rule packs from remote services.

## Single Service

Use this shape for one Go module with one owning team and a modest coverage
gate.

```yaml
version: 1

ownership:
  require_codeowners: true
  max_author_count_90d: 6

owners:
  aliases:
    "@service-team":
      - "alice@example.com"
      - "bob@example.com"
  modules:
    "example.com/acme/orders":
      owner: "@service-team"

coverage:
  min_package_coverage: 70

scoring:
  churn_max_lines_30d: 1000
  complexity_max_loc: 1000
  complexity_max_imports: 20
  complexity_max_files: 30
  dependency_centrality_max_reverse_imports: 10

suppression_policy:
  require_owner: true
  require_reason: true
  require_expiry: true
  max_days: 90
```

Run it with:

```sh
go test ./... -coverprofile=coverage.out
faultline scan ./... --coverage coverage.out --config faultline.yaml --format html --out faultline-report.html
```

## Multi-Module Monorepo

Use explicit module owners so repeated package suffixes and module moves do not
change ownership unexpectedly.

```yaml
version: 1

rule_packs:
  - path: .faultline/rules/platform.yaml

ownership:
  require_codeowners: true
  max_author_count_90d: 8

owners:
  aliases:
    "@platform-team":
      - "platform@example.com"
    "@checkout-team":
      - "checkout@example.com"
    "@shared-libraries":
      - "shared@example.com"
  modules:
    "example.com/acme/services/checkout":
      owner: "@checkout-team"
    "example.com/acme/services/billing":
      owner: "@platform-team"
    "example.com/acme/libs/shared":
      owner: "@shared-libraries"

coverage:
  min_package_coverage: 65

scoring:
  churn_max_lines_30d: 1500
  complexity_max_loc: 1500
  complexity_max_imports: 25
  complexity_max_files: 40
  dependency_centrality_max_reverse_imports: 15

suppression_policy:
  require_owner: true
  require_reason: true
  require_expiry: true
  max_days: 120
```

From the repository root, scan all modules:

```sh
faultline scan ./... --all-modules --config faultline.yaml --format json --out faultline-report.json
```

## CODEOWNERS-Heavy Repository

Use this shape when CODEOWNERS is the primary accountability source, but map
common authors or external handles back to canonical teams for clearer evidence.

```yaml
version: 1

ownership:
  require_codeowners: true
  max_author_count_90d: 6

owners:
  aliases:
    "@api-platform":
      - "api-maintainer@example.com"
      - "@github-org/api-reviewers"
    "@data-platform":
      - "data-maintainer@example.com"
      - "@github-org/data-reviewers"
  modules:
    "example.com/acme/api":
      owner: "@api-platform"

coverage:
  min_package_coverage: 75

suppression_policy:
  require_owner: true
  require_reason: true
  require_expiry: true
  max_days: 60
```

Validate CODEOWNERS and config together before scanning:

```sh
faultline config validate --config faultline.yaml --strict
faultline scan ./... --config faultline.yaml --strict-config --format sarif --out faultline.sarif
```

## Architecture Boundary Enforcement

Use boundaries to make import rules explicit. Boundary rules match package
import paths and package directories. Suppressions should be temporary waivers
with owners, reasons, creation dates, and expiries.

```yaml
version: 1

ownership:
  require_codeowners: true
  max_author_count_90d: 6

owners:
  modules:
    "example.com/acme/service":
      owner: "@service-team"

coverage:
  min_package_coverage: 70

boundaries:
  - name: handlers-must-not-import-storage
    from: "*/internal/handlers/*"
    deny:
      - "*/internal/storage/*"
    except:
      - "*/internal/storage/contracts"
  - name: domain-must-not-import-transport
    from: "*/internal/domain/*"
    deny:
      - "*/internal/http/*"
      - "*/internal/grpc/*"

suppression_policy:
  require_owner: true
  require_reason: true
  require_expiry: true
  max_days: 45

suppressions:
  - id: FL-BND-001
    category: BOUNDARY
    package: "*/internal/handlers/legacy"
    reason: "Temporary waiver while request mapping moves behind storage contracts"
    owner: "@service-team"
    created: "2026-04-30"
    expires: "2026-06-14"
```

Use `--fail-on high` when architecture violations should block CI:

```sh
faultline scan ./... --config faultline.yaml --strict-config --fail-on high --format json --out faultline-report.json
```
