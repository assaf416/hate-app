#!/usr/bin/env bash
# Runs the Hebrew Cucumber (godog) feature suite under features/.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

export CGO_ENABLED=1

echo "Running Cucumber (godog) scenarios from features/ ..."
go test -run TestFeatures -v .
