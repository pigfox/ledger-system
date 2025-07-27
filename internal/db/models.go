package db

import (
	"time"

	"github.com/google/uuid"
)

// User represents an application user.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserAddress maps a user's external wallet to a blockchain.
type UserAddress struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Chain   string `json:"chain"`   // e.g., "ethereum"
	Address string `json:"address"` // e.g., "0xabc..."
}

// Transaction represents a persisted transaction in the ledger system.
type Transaction struct {
	ID          uuid.UUID `json:"id"`
	UserID      int       `json:"user_id"`
	Type        string    `json:"type"` // e.g., "deposit", "withdrawal", "transfer", "reconciliation"
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"` // e.g., "ETH", "USDC"
	Status      string    `json:"status"`   // e.g., "pending", "completed"
	TxHash      string    `json:"tx_hash"`
	BlockHeight int64     `json:"block_height"`
	CreatedAt   time.Time `json:"created_at"`
}

// LedgerEntry is a double-entry bookkeeping row linked to a transaction.
type LedgerEntry struct {
	ID            string    `json:"id"`             // UUID
	TransactionID string    `json:"transaction_id"` // foreign key
	Account       string    `json:"account"`        // e.g., "user:123", "external"
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Direction     string    `json:"direction"` // "credit" or "debit"
	CreatedAt     time.Time `json:"created_at"`
}

// OnChainTransaction represents external blockchain data waiting for reconciliation.
type OnChainTransaction struct {
	ID          uuid.UUID `json:"id"`
	Address     string    `json:"address"` // external wallet
	TxHash      string    `json:"tx_hash"` // blockchain tx
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Direction   string    `json:"direction"` // "credit" or "debit"
	BlockHeight int64     `json:"block_height"`
	Reconciled  bool      `json:"reconciled"` // marked true after being matched
}

// ReconciliationReport summarizes the reconciliation results.
type ReconciliationReport struct {
	Matched      int                  `json:"Matched"`
	Flagged      int                  `json:"Flagged"`
	Errors       []string             `json:"Errors"`
	Incompatible []OnChainTransaction `json:"Incompatible"`
}

// TransactionRequest is used for incoming API payloads to create transactions.
type TransactionRequest struct {
	UserID      int     `json:"user_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`                   // e.g., "deposit", "withdrawal"
	TxHash      string  `json:"tx_hash,omitempty"`      // optional
	BlockHeight int     `json:"block_height,omitempty"` // optional
}

// TransferRequest is the request model for internal transfers.
type TransferRequest struct {
	FromUserID int     `json:"from_user_id"`
	ToUserID   int     `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

// Balance represents a user's token balance summary.
type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
