#!/bin/bash
set -e

echo "Building vault-reader..."

CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/vault-reader ./cmd/vault-reader

echo "Done: bin/vault-reader"
