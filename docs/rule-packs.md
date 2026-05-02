# Faultline Rule Packs

Rule packs are local, reusable Faultline policy files. They let a platform or security team publish common governance defaults while individual repositories keep their own suppressions and repo-specific overrides.

Faultline rule packs are local files only. Faultline does not fetch rule packs from a network, registry, GitHub repository, or SaaS endpoint.

For complete repository-level examples, see [config examples](config-examples.md).

## Schema

A repository config imports rule packs with `rule_packs`:

```yaml
version: 1

rule_packs:
  - path: .faultline/rules/platform.yaml
  - path: .faultline/rules/payments.yaml
```

Rule packs may contain:

```yaml
ownership:
  require_codeowners: true
  max_author_count_90d: 6

owners:
  aliases:
    "@payments-platform":
      - "alice@example.com"
      - "@github-team/payments"
  modules:
    "github.com/acme/service-a":
      owner: "@service-a-team"

coverage:
  min_package_coverage: 70

scoring:
  churn_max_lines_30d: 1500
  complexity_max_loc: 2500
  complexity_max_imports: 30
  complexity_max_files: 50
  dependency_centrality_max_reverse_imports: 15

boundaries:
  - name: handlers-must-not-import-storage
    from: "*/internal/handlers/*"
    deny:
      - "*/internal/storage/*"

suppression_policy:
  require_owner: true
  require_reason: true
  require_expiry: true
  max_days: 90
```

Rule packs must not contain `suppressions`. Suppressions are repo-local governance waivers and are ignored with a warning when found in a rule pack.

## Suppression Policy

Rule packs are the preferred place for organization-wide waiver rules:

```yaml
suppression_policy:
  require_owner: true
  require_reason: true
  require_expiry: true
  max_days: 90
```

The resolved `suppression_policy` is enforced against repo-local `suppressions` after all rule packs and repo overrides are merged. `max_days` limits waiver duration. If a suppression has `created: YYYY-MM-DD`, Faultline measures the maximum duration from that date. If `created` is absent, Faultline measures from the scan or config-validation date. Invalid `created` dates are reported and the scan date is used for the duration check.

In non-strict mode, a policy-violating suppression can still apply so existing CI does not unexpectedly fail, but the violation is included in config validation, config docs, suppression audits, scan reports, baseline outputs, PR reports, and SARIF metadata warnings. With `--strict-config`, policy warnings fail enforcement before scanning.

## Merge Model

Faultline resolves policy deterministically:

1. Start with Faultline defaults.
2. Apply rule packs in listed order.
3. Later rule packs override earlier rule packs for scalar settings, scoring calibration, owner aliases, and module owner entries.
4. Apply repo-local `faultline.yaml` overrides.
5. Append boundaries and de-duplicate by boundary `name`.
6. Apply repo-local suppressions after full policy resolution.

For duplicate boundary names:

- Identical duplicate rules are de-duplicated silently.
- Non-identical duplicate rules emit a warning.
- The later rule wins, so repo-local boundaries override rule-pack boundaries with the same name.

Owner aliases and module owner entries are map-like settings. Later rule packs override earlier entries with the same key, and repo-local `faultline.yaml` overrides imported entries. Scoring calibration is scalar policy and should be reviewed like any other enforcement threshold. Suppressions remain repo-local even when owner policy comes from a rule pack.

## Auditability

Resolved config outputs include:

- imported rule pack paths
- rule pack content hashes
- merge and validation warnings
- final resolved config hash

Use:

```sh
faultline config resolved --config faultline.yaml --format yaml --out resolved.yaml
faultline config docs --config faultline.yaml --format markdown --out faultline-policy.md
```

The resolved config hash is the enforcement artifact identity. Store `resolved.yaml` or `faultline-policy.md` as CI artifacts when policy traceability matters.

## Security Restrictions

Rule pack paths are treated as untrusted config input:

- Paths are local files only.
- Environment variable and shell expansion are not performed.
- Commands are never executed.
- Network fetching is not implemented.
- Paths must stay inside the repository root by default.
- Symlinks resolving outside the repository root are rejected.

Use `--allow-config-outside-repo` only for a trusted local policy checkout. That flag is intentionally explicit so CI reviewers can see when policy is loaded from outside the repository.

## Operational Risks

Rule packs centralize governance, so a bad pack can affect many repositories. Recommended controls:

- Version rule packs through normal code review.
- Pin rule-pack updates in repository pull requests.
- Run `faultline config validate --strict` in CI.
- Upload `faultline config resolved` output as a CI artifact.
- Keep suppressions repo-local so waiver ownership stays close to the affected code.
