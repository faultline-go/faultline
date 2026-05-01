# OSS Repository Settings

This checklist documents the GitHub repository configuration that supports a
credible open-source launch for Faultline.

## About Panel

Recommended values:

- Description: `Structural risk analysis for Go codebases`
- Website: `https://github.com/faultline-go/faultline#readme` until a dedicated
  docs site exists
- Topics:
  - `go`
  - `golang`
  - `cli`
  - `devtools`
  - `static-analysis`
  - `code-analysis`
  - `code-quality`
  - `architecture`
  - `risk-analysis`
  - `sarif`
  - `codeowners`
  - `github-actions`
  - `monorepo`
  - `governance`
  - `security-tools`

Enable:

- Issues
- Discussions
- Secret scanning
- Secret scanning push protection
- Dependabot alerts and security updates

Disable:

- Wiki, unless documentation moves out of the repository later
- Merge commits, if maintainers are comfortable with squash/rebase-only history

Projects should stay disabled until maintainers actively use GitHub Projects for
roadmap tracking. A stale empty project surface makes the repository look less
maintained.

Recommended pull request merge settings:

- allow squash merge
- allow rebase merge
- delete head branches on merge
- allow maintainers to update branches
- use pull request title and description for squash commit messages

## Branch Protection

Protect `main` with:

- require pull request before merging
- require status checks:
  - `test (ubuntu-latest)`
  - `test (macos-latest)`
  - `test (windows-latest)`
  - `dco`
- require branches to be up to date before merge
- require signed commits if the maintainer workflow supports it
- disallow force pushes
- disallow deletion

Do not require release dry-run checks on every PR; keep them manual or tag-only
unless release config becomes lightweight enough for routine validation.

## Releases

Use annotated semver tags:

```sh
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The release workflow should attach:

- Linux, macOS, and Windows archives
- `checksums.txt`
- Cosign checksum signature and certificate
- SPDX SBOM files
- GitHub provenance attestations

Mark early releases as pre-release until install, scan, SARIF, and Docker flows
have been validated by outside users.

## Packages

Primary container package:

```text
ghcr.io/faultline-go/faultline
```

After the first tagged release:

- make the GHCR package public
- connect the package to `faultline-go/faultline`
- keep immutable version tags
- keep `latest` for convenience only
- document digest pinning for production CI

The maintainer token used by `gh` must include `read:packages` to inspect
package state locally. GitHub Actions publishes with `GITHUB_TOKEN` and
`packages: write`.

## Labels

Seed these labels:

- `bug`
- `enhancement`
- `documentation`
- `dependencies`
- `go`
- `github-actions`
- `docker`
- `needs-triage`
- `good first issue`
- `help wanted`
- `breaking-change`

Labels should support triage and release notes, not mirror the entire product
taxonomy.

## Social Preview

Use `docs/assets/social-preview.svg` as the source artwork. GitHub's social
preview uploader currently expects PNG, JPG, or GIF, so export the SVG to a
1280x640 PNG before uploading it in repository settings.

Avoid marketing copy that implies cloud upload is required; the OSS trust
message should be local-first, source-free by default, and useful without login.
