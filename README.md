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

Save the partially populated .env.example as .env and fill in the required values.

## 🚀 Run

```bash

Terminal 1
./start.sh || ./shutdown.sh

Terminal 2
go run ./cmd/server

Terminal 3
./test.sh
```
See api.md for more details on the API.