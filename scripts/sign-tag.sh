#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
Usage: scripts/sign-tag.sh [--dry-run] vX.Y.Z [message]

Creates and verifies a signed annotated release tag using GPG loopback pinentry.

Environment:
  FAULTLINE_GPG_PASSPHRASE_FILE  Passphrase file. Default: /home/mike/.passphrase
  FAULTLINE_GPG_BIN              GPG binary. Default: gpg
USAGE
}

dry_run=false
if [ "${1:-}" = "--dry-run" ]; then
  dry_run=true
  shift
fi

tag="${1:-}"
message="${2:-Faultline ${tag}}"

if [ -z "${tag}" ]; then
  usage
  exit 2
fi

if ! [[ "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "expected semver tag like v1.2.3; got ${tag}" >&2
  exit 2
fi

passphrase_file="${FAULTLINE_GPG_PASSPHRASE_FILE:-/home/mike/.passphrase}"
if [ ! -r "${passphrase_file}" ]; then
  echo "passphrase file is not readable: ${passphrase_file}" >&2
  exit 2
fi

git rev-parse --git-dir >/dev/null

if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  echo "tag already exists locally: ${tag}" >&2
  exit 2
fi

if [ "${dry_run}" = true ]; then
  echo "would create signed tag ${tag}"
  echo "message: ${message}"
  echo "passphrase_file: ${passphrase_file}"
  exit 0
fi

wrapper="$(mktemp)"
chmod 700 "${wrapper}"
cleanup() {
  rm -f "${wrapper}"
}
trap cleanup EXIT

cat > "${wrapper}" <<'WRAPPER'
#!/usr/bin/env bash
set -euo pipefail
exec "${FAULTLINE_GPG_BIN:-gpg}" \
  --batch \
  --pinentry-mode loopback \
  --passphrase-file "${FAULTLINE_GPG_PASSPHRASE_FILE:?}" \
  "$@"
WRAPPER
chmod 700 "${wrapper}"

export FAULTLINE_GPG_PASSPHRASE_FILE="${passphrase_file}"
git -c gpg.program="${wrapper}" tag -s "${tag}" -m "${message}"
git tag -v "${tag}"

