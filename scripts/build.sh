#!/bin/bash
set -e

echo "Building vault-reader..."

mkdir -p bin
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/vault-reader ./cmd/vault-reader

echo "Done: bin/vault-reader"
