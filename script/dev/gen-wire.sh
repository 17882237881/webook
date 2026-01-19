#!/usr/bin/env bash
set -euo pipefail

# Generates Wire DI code for cmd/webook.
# Usage: bash script/dev/gen-wire.sh

go run -mod=mod github.com/google/wire/cmd/wire@v0.7.0 ./cmd/webook
