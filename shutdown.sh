#!/bin/bash
set -euo pipefail

set -x
set -e
clear
echo "Shutting down..."
docker compose down -v --remove-orphans && kill -9 $(lsof -ti tcp:8080) 2>/dev/null || true