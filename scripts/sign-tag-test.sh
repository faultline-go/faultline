#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script="${repo_root}/scripts/sign-tag.sh"

if [ ! -x "${script}" ]; then
  echo "scripts/sign-tag.sh must exist and be executable" >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

passphrase_file="${tmpdir}/passphrase"
printf 'test-passphrase\n' > "${passphrase_file}"
chmod 600 "${passphrase_file}"

output="$(
  FAULTLINE_GPG_PASSPHRASE_FILE="${passphrase_file}" \
    "${script}" --dry-run v1.2.3 2>&1
)"
case "${output}" in
  *"would create signed tag v1.2.3"*) ;;
  *)
    echo "dry-run output did not describe signed tag creation" >&2
    echo "${output}" >&2
    exit 1
    ;;
esac

invalid_output="${tmpdir}/sign-tag-invalid.out"
if FAULTLINE_GPG_PASSPHRASE_FILE="${passphrase_file}" "${script}" --dry-run 1.2.3 >"${invalid_output}" 2>&1; then
  echo "invalid tag without leading v should fail" >&2
  exit 1
fi
if ! grep -q "expected semver tag" "${invalid_output}"; then
  echo "invalid tag failure should explain expected semver format" >&2
  cat "${invalid_output}" >&2
  exit 1
fi
