# Release Process

Faultline releases are built with GoReleaser using pure-Go builds only. CGO is disabled for all release artifacts so users do not need a native SQLite or C toolchain.

## Prerequisites

- Go 1.26 or newer.
- GoReleaser installed locally for dry runs.
- Cosign installed if signing checksums locally.
- Syft installed if generating SBOMs locally.
- A clean working tree.
- Tags fetched locally.

## Local Verification

Run the standard checks before tagging:

```sh
go test ./...
go vet ./...
CGO_ENABLED=0 go build ./cmd/faultline
make build-all
make checksums
goreleaser release --snapshot --clean
```

## v1.0 Release Candidate Gate

Before promoting a v1.0 release candidate, run the OSS dogfood and release
contract gate from a clean working tree:

```sh
make quality
go test ./... -coverprofile=/tmp/faultline-cover.out
go run ./cmd/faultline config validate --config .faultline.yaml --strict
go run ./cmd/faultline scan ./... \
  --coverage /tmp/faultline-cover.out \
  --config .faultline.yaml \
  --format json \
  --out /tmp/faultline-self.json \
  --no-history
go run ./cmd/faultline scan ./... \
  --coverage /tmp/faultline-cover.out \
  --config .faultline.yaml \
  --format sarif \
  --out /tmp/faultline-self.sarif \
  --no-history
go run ./cmd/faultline scan ./... \
  --coverage /tmp/faultline-cover.out \
  --config .faultline.yaml \
  --format snapshot \
  --out /tmp/faultline-self-snapshot.json \
  --no-history
goreleaser release --snapshot --clean
docker build -t faultline:local .
docker run --rm faultline:local version
```

The gate proves the repository config is current, the OSS scanner can consume
its own coverage profile, JSON/SARIF/snapshot exports are emitted, and the
release packaging and container paths still work locally.

`make build-all` writes platform binaries to `dist/` for:

- linux amd64/arm64
- darwin amd64/arm64
- windows amd64/arm64

Unsigned local snapshot releases are expected to work without Cosign, Syft, GitHub credentials, or OIDC. Signing and provenance are release-workflow concerns, not local build requirements.

Validate the container image locally when Docker is available:

```sh
docker build -t faultline:local .
docker run --rm faultline:local version
```

Optional local supply-chain checks:

```sh
make sbom
make checksums
make sign-checksums
```

`make sbom` requires Syft and writes SPDX JSON SBOMs beside release archives. `make sign-checksums` requires Cosign and signs `dist/checksums.txt`.

## Tagging

Faultline uses semver tags:

```sh
VERSION=vX.Y.Z
bash scripts/sign-tag.sh "${VERSION}"
git tag -v "${VERSION}"
git push origin "${VERSION}"
```

`scripts/sign-tag.sh` creates a signed annotated tag through GPG loopback
pinentry and verifies the tag before returning. It reads the passphrase from
`/home/mike/.passphrase` by default; set `FAULTLINE_GPG_PASSPHRASE_FILE` to use
another local secret file. Run `make sign-tag-test` after editing the helper or
release docs.

The CI release dry run runs on manual workflow dispatch without signing or provenance. Version tags run the release workflow, which builds artifacts, generates SBOMs, refreshes checksums, signs the checksum file with Cosign keyless signing, publishes the GHCR image when registry permissions are available, requests GitHub artifact attestations, and uploads release assets to the GitHub release.

GoReleaser's direct GitHub release pipe is intentionally disabled. The workflow
creates or updates the GitHub release after SBOMs, refreshed checksums,
signatures, and attestations are available so users see a complete asset set.

## Homebrew

GoReleaser is configured to generate a Homebrew cask named `faultline` for the tap repo `faultline-go/homebrew-tap`. Faultline publishes through `homebrew_casks`, which is the current GoReleaser path for prebuilt CLI binaries.

The cask includes:

- description and homepage
- Apache-2.0 license
- binary stanza for the `faultline` executable

Tap publishing is token-gated. Set `HOMEBREW_TAP_GITHUB_TOKEN` in the release workflow environment to allow GoReleaser to push cask updates. Without that token, cask upload is skipped.

Expected user install path after the tap exists:

```sh
brew tap faultline-go/tap
brew install --cask faultline
faultline version
```

## Docker Images

The release workflow builds and publishes multi-arch images to:

```text
ghcr.io/faultline-go/faultline
```

Configured platforms:

- linux/amd64
- linux/arm64

Tags:

- `vX.Y.Z`
- `X.Y`
- `latest`

The runtime image is based on Alpine Go and includes the Faultline binary, Go, git, CA certificates, README, LICENSE, and the example config. Go and git are included because package loading and git history analysis need those tools for scans in CI.

GHCR publishing assumes:

- the release workflow has `packages: write`
- the workflow can log in with `GITHUB_TOKEN`
- the `ghcr.io/faultline-go/faultline` package namespace is allowed for the repository

Snapshot and local GoReleaser dry runs disable Docker publishing.

Container users should verify image provenance through GHCR/GitHub metadata and pinned immutable digests where possible:

```sh
docker pull ghcr.io/faultline-go/faultline:vX.Y.Z
docker image inspect ghcr.io/faultline-go/faultline:vX.Y.Z
```

Container tags are convenient, but production CI should prefer pinned digests after the image has been verified.

## Version Metadata

Release builds inject:

- `internal/version.Version`
- `internal/version.Commit`
- `internal/version.BuildDate`

Verify a binary:

```sh
faultline version
```

The output includes version, commit, build date, Go version, and OS/architecture.

## Checksum Verification

Checksums provide integrity: they show that the file you downloaded matches the file listed in `checksums.txt`.

Every release should include `checksums.txt`. Verify archives and SBOMs before installation:

```sh
sha256sum --check --ignore-missing checksums.txt
```

macOS users can use:

```sh
shasum -a 256 -c checksums.txt
```

Windows users can compare with:

```powershell
Get-FileHash .\faultline_X.Y.Z_windows_amd64.zip -Algorithm SHA256
```

Checksums alone do not prove who produced the checksum file. Use signature verification for authenticity.

## Signature Verification

Faultline signs `checksums.txt`, not every archive individually. Because the signed checksum file covers the archives and SBOMs, this keeps verification simple and auditable.

Verify the checksum signature with Cosign:

```sh
cosign verify-blob checksums.txt \
  --signature checksums.txt.sig \
  --certificate checksums.txt.pem \
  --certificate-identity-regexp 'https://github.com/faultline-go/faultline/.github/workflows/release.yml@refs/tags/v.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

Signature verification proves that the checksum file was signed by the Faultline GitHub Actions release workflow using GitHub OIDC identity. It does not prove that the source code was safe or bug-free.

## SBOM Inspection

Release workflows generate SPDX JSON SBOMs for release archives:

```text
faultline_X.Y.Z_linux_amd64.tar.gz.spdx.json
faultline_X.Y.Z_windows_amd64.zip.spdx.json
```

Inspect an SBOM with `jq`:

```sh
jq '.packages[] | {name: .name, versionInfo: .versionInfo}' faultline_X.Y.Z_linux_amd64.tar.gz.spdx.json
```

The SBOM describes packaged software contents and dependency metadata. It is not a vulnerability scan and does not assert that dependencies are risk-free.

## Provenance Verification

Release workflows request GitHub artifact attestations for archives, SBOMs, and `checksums.txt`.

Verify provenance with the GitHub CLI:

```sh
gh attestation verify faultline_X.Y.Z_linux_amd64.tar.gz --repo faultline-go/faultline
```

Provenance links an artifact to the GitHub Actions workflow that built it. It does not replace checksum or signature verification; use all three for stronger assurance:

1. Verify the checksum for file integrity.
2. Verify the Cosign signature for checksum authenticity.
3. Verify provenance for build workflow identity.

## Rollback Guidance

If a release is bad:

1. Mark the GitHub release as pre-release or delete the release assets if they should not be used.
2. Leave the git tag in place unless the release never became public.
3. Cut a new patch release with the fix, for example `v0.1.1`.
4. Document the issue and recommended upgrade path in the changelog.

Faultline does not yet publish Kubernetes manifests, Helm charts, enterprise PKI, or SLSA level guarantees beyond GitHub artifact attestations.
