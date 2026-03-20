#!/bin/bash
set -e

BUILD_VERSION="${BUILD_VERSION:-dev}"
BUILD_GITHASH="${BUILD_GITHASH:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_TIME="${BUILD_TIME:-$(date -u '+%Y-%m-%dT%H:%M:%SZ')}"
BRANCH="${BRANCH:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)}"

echo "Building version=${BUILD_VERSION} hash=${BUILD_GITHASH} branch=${BRANCH}"

CGO_ENABLED=0 go build \
  -ldflags "-X main.version=${BUILD_VERSION} -X main.gitHash=${BUILD_GITHASH} -X main.buildStamp=${BUILD_TIME} -X main.branch=${BRANCH}" \
  -o server .
