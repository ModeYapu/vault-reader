#!/bin/bash
set -e

VERSION="${1:-dev}"

echo "Building vault-reader ${VERSION}..."

mkdir -p bin
CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o bin/vault-reader ./cmd/vault-reader

echo "Done: bin/vault-reader"
