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
./start.sh || ./shutdown.sh # Shutdown will remove everything in the database

Terminal 2
go run ./cmd/server

Terminal 3
./test.sh

# Cheat script to run all API calls in sequence
Terminal 4
./run-sequence.sh
```
See api.md for more details on the API.

The following address have been tested:
https://etherscan.io/address/0xdadb0d80178819f2319190d340ce9a924f783711
https://etherscan.io/address/0x0013d9bb14d37654cdacb0a00209c6994511afa1

# Difference Between `transactions` and `onchain_transactions`

| Table                | Purpose                                                                | Source                    | Used In        |
|----------------------|------------------------------------------------------------------------|---------------------------|----------------|
| `onchain_transactions` | Represents raw transactions fetched from a blockchain (ETH, MATIC, etc.) | On-chain via indexers or RPC | Reconciliation |
| `transactions`         | Internal ledger activity: deposits, withdrawals, transfers, or reconciled entries | Internal system            | Ledger logic    |


To view raw data in the database, you can use the following commands:

```bash
# psql "postgres://test_xyz:test_xyzpass@localhost:55432/test_xyzledger"
# psql "postgres://xyz:xyzpass@localhost:5432/xyzledger"