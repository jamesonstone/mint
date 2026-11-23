#!/usr/bin/env bash
set -euo pipefail

binary=$1
notes=$2
: "${VERSION:?}" "${STAGING_DIGEST:?}" "${GITHUB_SHA:?}" "${GITHUB_REPOSITORY:?}"

digest=$(bash "$(dirname "$0")/verify-release-binary.sh" "$binary" "$STAGING_DIGEST" "$VERSION")
"$binary" release tag --tag "$VERSION" --target "$GITHUB_SHA" --notes-file "$notes"

find_release() {
  gh api --paginate --slurp "repos/$GITHUB_REPOSITORY/releases?per_page=100" |
    jq -c --arg tag "$VERSION" '
      [.[][] | select(.tag_name == $tag)] |
      if length > 1 then error("multiple releases for the version") else first // null end'
}

release=$(find_release)
if [[ "$release" == null ]]; then
  gh release create "$VERSION" --repo "$GITHUB_REPOSITORY" --draft --verify-tag \
    --target "$GITHUB_SHA" --title "$VERSION" --notes-file "$notes"
  release=$(find_release)
fi
release_id=$(jq -er '.id' <<< "$release")
draft=$(jq -r '.draft' <<< "$release")

asset_name=$(basename "$binary")
find_asset_digest() {
  gh api --paginate --slurp "repos/$GITHUB_REPOSITORY/releases/$release_id/assets?per_page=100" |
    jq -r --arg name "$asset_name" '
      [.[][] | select(.name == $name)] |
      if length == 0 then ""
      elif length == 1 and .[0].state == "uploaded" then .[0].digest
      else error("release asset is ambiguous or incomplete") end'
}

observed=$(find_asset_digest)
if [[ -z "$observed" && "$draft" == true ]]; then
  gh release upload "$VERSION" "$binary" --repo "$GITHUB_REPOSITORY"
  observed=$(find_asset_digest)
fi
"$binary" verify --expected "$digest" --observed "$observed"
if [[ "$draft" == true ]]; then
  gh release edit "$VERSION" --repo "$GITHUB_REPOSITORY" --draft=false
fi
published=$(find_release)
jq -e --argjson id "$release_id" '.id == $id and .draft == false and .prerelease == false and .published_at != null' \
  <<< "$published" >/dev/null
"$binary" verify --expected "$digest" --observed "$(find_asset_digest)"
printf 'Published %s with binary digest %s\n' "$VERSION" "$digest"
