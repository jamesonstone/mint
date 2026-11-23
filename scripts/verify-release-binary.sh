#!/usr/bin/env bash
set -euo pipefail

binary=$1
expected=$2
version=$3

checksum=$(shasum -a 256 "$binary")
observed="sha256:${checksum%% *}"
# Check the bytes independently before executing the downloaded program.
if [[ "$observed" != "$expected" ]]; then
  printf 'binary digest mismatch: expected %s, observed %s\n' "$expected" "$observed" >&2
  exit 1
fi
chmod +x "$binary"
"$binary" verify --expected "$expected" --observed "$observed"
actual_version=$("$binary" version)
if [[ "$actual_version" != "$version" ]]; then
  printf 'binary version mismatch: expected %s, observed %s\n' "$version" "$actual_version" >&2
  exit 1
fi
printf '%s\n' "$observed"
