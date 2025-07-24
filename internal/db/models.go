package db

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserAddress struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

type TransactionRequest struct {
	UserID   int     `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	TxHash   string  `json:"tx_hash,omitempty"`
	Type     string  `json:"type"`
}

type TransferRequest struct {
	FromUserID int     `json:"from_user_id"`
	ToUserID   int     `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type OnChainTransaction struct {
	ID        uuid.UUID
	Address   string
	TxHash    string
	Amount    float64
	Currency  string
	Direction string // "credit" or "debit"
}

type LedgerEntry struct {
	ID            string    // UUID, typically generated in code
	TransactionID string    // foreign key to transactions.id
	Account       string    // e.g., "user:123", "external"
	Amount        float64   // stored as NUMERIC in DB
	Currency      string    // e.g., "ETH", "USDC"
	Direction     string    // "credit" or "debit"
	CreatedAt     time.Time // auto-filled by DB
}

type ReconciliationReport struct {
	Matched      int
	Flagged      int
	Errors       []string
	Incompatible []OnChainTransaction
}
