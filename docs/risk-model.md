# Faultline Risk Model

Faultline's current scoring version is `faultline-risk-v0.2`. Scores are reported on a 0-100 scale. The arithmetic is deterministic for a fixed scan input, but some inputs are time-windowed. In particular, churn and author counts are read from `git log --since="30 days ago"` and `git log --since="90 days ago"`, so the same repository can legitimately score differently as commits age out of those windows.

```text
risk_score =
  churn_score * 0.25 +
  coverage_gap_score * 0.20 +
  complexity_score * 0.20 +
  ownership_entropy_score * 0.20 +
  dependency_centrality_score * 0.15
```

## Component Scores

- `churn_score`: added plus deleted lines from git history in the last 30 days, capped at 100.
- `coverage_gap_score`: distance below configured minimum package coverage. Unknown coverage uses a neutral score of 50.
- `complexity_score`: simple LOC, non-standard import count, and non-generated file count blend.
- `ownership_entropy_score`: normalized diversity of recent git authors.
- `dependency_centrality_score`: reverse import count across loaded packages. Standard library imports are excluded.

## Findings

Current rules:

- `FL-CHURN-001`: high churn in the last 30 days.
- `FL-COV-001`: low coverage where coverage is known.
- `FL-COV-002`: missing coverage data, informational only.
- `FL-OWN-001`: no owner found when ownership is required.
- `FL-OWN-002`: high author count in the last 90 days.
- `FL-OWN-003`: CODEOWNERS owner differs from the dominant git author or configured alias.
- `FL-OWN-004`: module owner missing in a multi-module repository.
- `FL-DEP-001`: high reverse import count.
- `FL-DEP-002`: local `replace` directive present.
- `FL-DEP-003`: module replaced to a different module path.
- `FL-DEP-004`: direct dependency appears unused, best-effort.
- `FL-DEP-005`: dependency has unusually broad blast radius based on imports.
- `FL-DEP-006`: dependency version is a pseudo-version.
- `FL-DEP-007`: local replace points to another module in the same repository.
- `FL-GEN-001`: generated-code-heavy package.
- `FL-BND-001`: configured architecture boundary violation.

Every finding includes severity, description, evidence, recommendation, and confidence.

Boundary findings are produced from `faultline.yaml` policy rules. They are high severity by default and include the exact importing package, exact denied import, rule name, `from` pattern, `deny` pattern, and matched import evidence. Boundary findings are separate from the numeric risk score in this MVP.

## Dependency Risk

Dependency risk is structural module-governance analysis, not CVE scanning. Faultline reads local `go.mod` and `go.sum`, then compares required modules with loaded package imports. It reports direct versus indirect requirements, replacement directives, local replacement paths, cross-module local replaces inside a repository, pseudo-versions, best-effort unused direct requirements, and dependencies imported by many packages.

Faultline does not call the Go module proxy, run `go get`, mutate `go.mod` or `go.sum`, or query vulnerability databases during default dependency analysis. The optional `--govulncheck auto|/path/to/govulncheck` mode runs an external `govulncheck` binary only when explicitly requested. That output is labeled as external tool output and should be governed separately from Faultline's structural dependency findings.

Dependency findings are currently top-level report findings and are not folded into the package risk score. This avoids mixing module governance signals with package-local structural risk until enough real repository data exists to tune weights responsibly.

## Calibration Notes

The initial model is intentionally simple and should be treated as a prioritization heuristic. Several constants are deliberately conservative until Faultline has real-world calibration data:

- Churn reaches the maximum component score at 1,000 added plus deleted lines in the last 30 days.
- LOC complexity reaches its maximum LOC subcomponent at 1,000 non-generated lines.
- Unknown coverage is scored as 50 on the coverage-gap component. Because coverage has a 0.20 weight, a package with no other risk signals receives 10 risk-score points when coverage is unknown. This is not the same as assigning 0% coverage; it is a bias toward making missing coverage visible without making it a failure by itself.

These thresholds are stable and explainable, but they are not yet statistically calibrated. Large refactors, generated-code churn, or repos without coverage profiles can shift many packages upward at once. Enterprise deployments should compare scores within the same repository and CI setup before using absolute thresholds as hard gates.

The calibration constants are configurable under `scoring` in `faultline.yaml` or imported rule packs. Defaults are:

```yaml
scoring:
  churn_max_lines_30d: 1000
  complexity_max_loc: 1000
  complexity_max_imports: 20
  complexity_max_files: 30
  dependency_centrality_max_reverse_imports: 10
```

Changing these values changes normalized component scores while preserving the same weighted formula. Treat those changes as policy changes: review them, commit them, and use `faultline config explain` or `faultline config docs` to make the resolved calibration auditable.

See [config examples](config-examples.md) for sanitized starting points that tune these values for common repository shapes.

## Monorepos And Workspaces

Faultline discovers multiple `go.mod` files under the repository root and detects a root `go.work` when present. Reports include module path, module root, `go.mod` path, go.work inclusion, and whether the module was selected for a scan.

From the repository root, `faultline scan ./...` scans all discovered modules. From inside a module, it scans that module unless `--all-modules` is supplied. `--module` and `--ignore-module` select modules by module path or module-root path. Package records include module identity so repeated import paths in different modules have stable package identities.

## Ownership Signals

Faultline resolves owners from local, auditable inputs in this order:

1. explicit module owner from `owners.modules`
2. CODEOWNERS
3. dominant git author mapped through `owners.aliases`
4. dominant git author as a low-confidence fallback
5. unknown

All candidate owners are reported as evidence. The selected owner includes an owner source and confidence so downstream governance can distinguish explicit ownership from git-authorship inference. In multi-module repositories, `owners.modules` entries are expected for each module; missing entries can produce `FL-OWN-004`.

Package ownership is the package-level signal used in risk reports. File ownership is resolved separately from CODEOWNERS only where exact file context matters: PR changed-file summaries, boundary findings with an importing file, and SARIF results that point to a concrete file. This keeps full scan reports compact while preserving reviewer-specific ownership evidence.

`owners.aliases` maps emails or external team handles to canonical owner teams:

```yaml
owners:
  aliases:
    "@payments-platform":
      - "alice@example.com"
      - "@github-team/payments"
  modules:
    "github.com/acme/service-a":
      owner: "@service-a-team"
```

Git authorship is a weak signal. It can identify likely maintainers when CODEOWNERS is missing, but it can be distorted by large refactors, bot commits, pair-programming commits, imported history, or recent incident work. Treat `FL-OWN-003` as a review prompt to reconcile responsibility, not proof that CODEOWNERS is wrong.

## CODEOWNERS Compatibility

Faultline treats CODEOWNERS as governance evidence. It searches `.github/CODEOWNERS`, `CODEOWNERS`, and `docs/CODEOWNERS` in that order, then applies GitHub's last-match-wins rule within the selected file. Ownership evidence includes the matched CODEOWNERS file, line number, pattern, and owners from the selected rule.

For boundary findings, Faultline locates the importing file on a best-effort basis and attaches file-level CODEOWNERS owners, matched rule, and line metadata. PR reviews summarize changed files with owners, changed files without owners, and mismatches between file-level owners and package owners. SARIF results with exact locations include file owner metadata in result properties.

Supported behavior includes comments, blank lines, multiple owners, root-anchored patterns, directory patterns, common wildcard patterns, and escaped spaces in patterns where practical. Faultline reports warnings for malformed rules, rules without owners, unsupported patterns, and owner tokens that do not start with `@` and are not email-like.

Known deviations from GitHub semantics:

- Matching is package-oriented and approximate, while GitHub evaluates ownership for specific files.
- Unsupported CODEOWNERS constructs such as `!` negation and character ranges are reported as warnings so governance reviewers can fix them.
- Faultline does not contact GitHub to validate users, teams, or organization membership.
- CODEOWNERS evidence is not proof of current accountability; it is one ownership input alongside module owners and git authorship.

## Assumptions

Faultline assumes package-level signals are useful for prioritization. It treats git history, CODEOWNERS, and coverage profiles as optional local inputs. Missing inputs produce warnings and evidence rather than scan failures.

Generated files are counted separately. By default, generated LOC does not contribute to LOC complexity. Use `--include-generated` only when generated code should be included in complexity scoring.

## False-Positive Risks

- Large stable packages can score high because many packages import them.
- Shallow clones can understate churn and author diversity.
- Coverage profiles that do not include all packages can make coverage appear unknown.
- CODEOWNERS patterns are approximated and may differ from GitHub in edge cases.
- Git authorship can reflect recent activity rather than durable stewardship.
- Generated-heavy packages may look structurally risky even when the generator is the real maintenance boundary.
- Dependency usage is based on loaded import paths. Build tags, generated code, dynamic plugin loading, or tool-only modules can make unused-dependency detection conservative or noisy.

## Suppressions

Suppressions are waivers, not deletions. A matching active suppression marks a finding as suppressed and adds owner, reason, package pattern, optional category, created date, and expiry metadata when supplied. The finding remains in package findings and also appears in the top-level suppression audit.

Suppressed findings are excluded from `--fail-on` behavior and unsuppressed severity counts. Expired suppressions are ignored. Suppressions missing always-required identity fields (`id` and `package`) are not applied. `reason`, `owner`, and `expires` are required when the resolved `suppression_policy` requires them.

`suppression_policy.max_days` limits waiver duration. Faultline measures the limit from `suppressions[].created` when present, otherwise from the scan or config-validation date. Policy-violating suppressions are reported in non-strict mode and may still apply; `--strict-config` fails before enforcement so CI can require clean waiver governance.

Governance expectation: every suppression must have a human or team owner, reason, creation date, and expiry date so waivers are revisited. `faultline suppressions audit` reports active, expired, expiring, incomplete, policy-violating, and currently unmatched suppressions. Unmatched suppressions are important because they can indicate stale waivers that no longer correspond to active findings.

## Baselines

Baselines are governance snapshots for ratcheting. `faultline baseline create` stores scan metadata, repository fingerprint, config hash, package risk scores, component score breakdowns, stable finding identities, active suppression metadata, and summary counts. It does not store source code or file contents.

`faultline baseline check` compares the current scan against that snapshot and reports:

- new unsuppressed findings
- resolved findings
- worsened packages
- improved packages
- currently suppressed findings

Finding identity is intentionally source-free. It is based on finding ID, package import path, category, stable evidence key/value pairs, boundary denied import evidence when present, and package-level location metadata where available.

Baseline gates are explicit:

- `--fail-on-new high|critical|none` fails only on new unsuppressed findings at or above the selected severity.
- `--fail-on-risk-delta <number>` fails when a package risk score increases by more than the configured threshold.

Suppressions remain waivers during baseline checks. Suppressed findings are listed for auditability, but they do not fail `--fail-on-new`. A baseline should be treated as accepted debt, not proof that the code is safe. Teams should review and refresh baselines only through an intentional governance process.

## Config Governance

`faultline config validate` and `faultline config explain` make policy inputs auditable. Validation warnings cover malformed boundary rules, incomplete or expired suppressions, suspicious thresholds, invalid expiry dates, and unknown top-level keys. `--strict-config` on scan, baseline, and PR commands turns those warnings into execution failures so CI cannot silently enforce a questionable policy file.

All report formats include the config hash where practical, allowing a scan artifact to be tied back to the exact policy input used during analysis.
