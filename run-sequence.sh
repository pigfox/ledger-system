#!/bin/bash
set -euo pipefail

if [[ ! -f .env ]]; then
  echo "❌ .env file not found"
  exit 1
fi

set -x
set -e
clear
echo "Running all API calls"

API_URL="http://localhost:8080"
API_KEY="qwerty123456"
HEADERS=(-H "X-API-Key: $API_KEY" -H "Content-Type: application/json")

# Health check (no headers)
curl "$API_URL/health"

# Create users
curl -X POST "${API_URL}/api/v1/users" "${HEADERS[@]}" -d '{"name": "user", "email": "user@example.com"}'
curl -X POST "${API_URL}/api/v1/users" "${HEADERS[@]}" -d '{"name": "user2", "email": "user2@example.com"}'

# Deposits
curl -X POST "${API_URL}/api/v1/transactions/deposit" "${HEADERS[@]}" -d '{"user_id": 1, "amount": 13.5, "currency": "ETH"}'
curl -X POST "${API_URL}/api/v1/transactions/deposit" "${HEADERS[@]}" -d '{"user_id": 1, "amount": 3.6, "currency": "USDC"}'
curl -X POST "${API_URL}/api/v1/transactions/deposit" "${HEADERS[@]}" -d '{"user_id": 1, "amount": 4.4, "currency": "MATIC"}'

# Withdraw
curl -X POST "${API_URL}/api/v1/transactions/withdraw" "${HEADERS[@]}" -d '{"user_id": 1, "amount": 2.5, "currency": "ETH"}'

# Transfer
curl -X POST "${API_URL}/api/v1/transactions/transfer" "${HEADERS[@]}" -d '{"from_user_id": 1, "to_user_id": 2, "amount": 1.5, "currency": "USDC"}'

# Get user balances
curl "${API_URL}/api/v1/users/1/balances" "${HEADERS[@]}"
curl "${API_URL}/api/v1/users/1/balances?currency=ETH" "${HEADERS[@]}"
curl "${API_URL}/api/v1/users/1/balances?currency=MATIC" "${HEADERS[@]}"
curl "${API_URL}/api/v1/users/1/balances?currency=USDC" "${HEADERS[@]}"

# Add addresses (adjust if your schema needs more fields)
curl -X POST "${API_URL}/api/v1/addresses" "${HEADERS[@]}" -d '{"user_id": 1, "chain": "ethereum","address": "0xdadb0d80178819f2319190d340ce9a924f783711"}'
curl -X POST "${API_URL}/api/v1/addresses" "${HEADERS[@]}" -d '{"user_id": 1, "chain": "ethereum", "address": "0x0013d9bb14d37654cdacb0a00209c6994511afa1"}'
curl -X POST "${API_URL}/api/v1/addresses" "${HEADERS[@]}" -d '{"user_id": 2, "chain": "ethereum", "address": "0x9c1becd2f442e86a326032edbb644f38ce8bce96"}'

# Get address transactions
curl "${API_URL}/api/v1/addresses/0xdadb0d80178819f2319190d340ce9a924f783711/transactions" "${HEADERS[@]}"

# Get address balances
curl "${API_URL}/api/v1/addresses/0xdadb0d80178819f2319190d340ce9a924f783711/balances" "${HEADERS[@]}"

# Reconciliation
curl -X POST "${API_URL}/api/v1/reconciliation" "${HEADERS[@]}"
