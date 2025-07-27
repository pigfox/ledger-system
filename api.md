# 📘 Ledger System API Documentation

All endpoints (except `/health`) require an `X-API-Key` header. JSON body requests must also set `Content-Type: application/json`.

## 🔐 Authentication Headers

```
X-API-Key: your-api-key
Content-Type: application/json
```

---

## 🩺 Health

### `GET /health`

- ✅ No API key required
- 🔍 Returns service status

**Response:**
```text
OK
```

---

## 👤 Users

### `POST /api/v1/users`

- 📄 Create a new user

**Request:**
```json
{
  "name": "Alice",
  "email": "alice@example.com"
}
```

**Response:**
```json
{
  "id": 1
}
```

---

## 💰 Transactions

### `POST /api/v1/transactions/deposit`

- 💸 Deposit funds into a user account

**Request:**
```json
{
  "user_id": 1,
  "amount": 13.5,
  "currency": "ETH"
}
```

**Response:**
```json
{
  "id": "uuid"
}
```

---

### `POST /api/v1/transactions/withdraw`

- 🏧 Withdraw funds from a user account

**Request:**
```json
{
  "user_id": 1,
  "amount": 2.5,
  "currency": "ETH"
}
```

**Response:**
```json
{
  "id": "uuid"
}
```

---

### `POST /api/v1/transactions/transfer`

- 🔁 Transfer funds between users

**Request:**
```json
{
  "from_user_id": 1,
  "to_user_id": 2,
  "amount": 1.5,
  "currency": "USDC"
}
```

**Response:**
```json
{
  "id": "uuid"
}
```

---

## 📊 Balances

### `GET /api/v1/users/{userId}/balances`

- 🧾 Get all balances for a user

**Example:**
```
GET /api/v1/users/1/balances
```

**Response:**
```json
[
  {
    "currency": "ETH",
    "balance": 10.5
  },
  {
    "currency": "USDC",
    "balance": 7.0
  }
]
```

---

### `GET /api/v1/users/{userId}/balances?currency=ETH`

- 📈 Get balance for a specific currency

**Example:**
```
GET /api/v1/users/1/balances?currency=ETH
```

---

## ⛓️ Blockchain Addresses

### `POST /api/v1/addresses`

- 🔗 Register a blockchain address

**Request:**
```json
{
  "user_id": 1,
  "chain": "ethereum",
  "address": "0xdadb0d80178819f2319190d340ce9a924f783711"
}
```

**Response:**
```json
{
  "user_id": 1,
  "chain": "ethereum",
  "address": "0xdadb0d80178819f2319190d340ce9a924f783711"
}
```

---

### `GET /api/v1/addresses/{address}/transactions`

- 📜 Get on-chain transactions for an address

**Example:**
```
GET /api/v1/addresses/0xdadb0.../transactions
```

**Response:**
```json
[
  {
    "id": "uuid",
    "tx_hash": "0x...",
    "amount": 3.4,
    "currency": "ETH",
    "direction": "credit",
    "block_height": 123456,
    "created_at": "2025-07-26T00:00:00Z"
  }
]
```

---

### `GET /api/v1/addresses/{address}/balances`

- 📦 Get current on-chain balances for an address

**Response:**
```json
{
  "ETH": 3.2,
  "USDC": 7.1,
  "MATIC": 0.4
}
```

---

## 🔄 Reconciliation

### `POST /api/v1/reconciliation`

- 🔄 Reconcile on-chain and ledger entries

**Response:**
```json
{
  "Matched": 1,
  "Flagged": 0,
  "Errors": [],
  "Incompatible": []
}
```
