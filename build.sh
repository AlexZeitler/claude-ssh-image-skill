#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "Building ccimgd (daemon)..."
cd daemon
CGO_ENABLED=0 go build -o ccimgd .
echo "  -> daemon/ccimgd"

echo "Building ccimg (client)..."
cd ../client
CGO_ENABLED=0 go build -o ccimg .
echo "  -> client/ccimg"

echo "Done."
