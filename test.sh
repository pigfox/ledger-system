#!/bin/bash
set -euo pipefail

if [[ ! -f .env ]]; then
  echo "❌ .env file not found"
  exit 1
fi

set -x
set -e
clear
echo "Running tests..."
source .env && TESTING=true go test -v -count=1 ./test/...