#!/usr/bin/env bash

set -euo pipefail

version="${1:-}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Expected a stable semantic version, received: $version" >&2
  exit 1
fi

if [[ -z "${GITHUB_REPOSITORY:-}" ]]; then
  echo "GITHUB_REPOSITORY must be set to publish a release image" >&2
  exit 1
fi

registry="${REGISTRY:-ghcr.io}"
image_name="${IMAGE_NAME:-${GITHUB_REPOSITORY,,}}"
image="${registry}/${image_name}"
image="${image,,}"
major="${version%%.*}"
minor_and_patch="${version#*.}"
minor="${minor_and_patch%%.*}"

docker buildx build \
  --file Dockerfile \
  --platform linux/amd64,linux/arm64 \
  --progress plain \
  --push \
  --provenance=mode=max \
  --sbom=true \
  --label "org.opencontainers.image.source=https://github.com/${GITHUB_REPOSITORY}" \
  --label "org.opencontainers.image.version=v${version}" \
  --tag "${image}:v${version}" \
  --tag "${image}:v${major}.${minor}" \
  --tag "${image}:v${major}" \
  --tag "${image}:latest" \
  .
