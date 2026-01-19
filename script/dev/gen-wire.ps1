# Generates Wire DI code for cmd/webook.
# Usage: powershell -ExecutionPolicy Bypass -File script\\dev\\gen-wire.ps1
$ErrorActionPreference = 'Stop'

go run -mod=mod github.com/google/wire/cmd/wire@v0.7.0 ./cmd/webook
