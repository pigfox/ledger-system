# ledger-system
A ledger system


A Go-based internal ledger + blockchain reconciler supporting ETH, MATIC, and USDC across Ethereum and Polygon.

---

## 🧪 Features

- Double-entry ledger: tracks deposits, withdrawals, and transfers
- Blockchain indexer: scans ETH & ERC-20 transactions
- Reconciliation engine: compares ledger and chain state
- RESTful API: exposed via Gorilla Mux
- Added features:
  - Api key authentication
  - Api versioning
  - Health endpoint

---

## 🚀 Run

```bash
docker-compose up / docker-compose down
go run ./cmd/server

go clean -testcache && go test -v -count=1 ./test/...