package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var Conn Connection

type Connection struct {
	DB *sql.DB
}

func init() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("DB connect error:", err)
	}
	Conn.DB = db
}

func Connect() Connection {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	Conn.DB = db
	return Conn
}

// USERS
func CreateUser(u User) (User, error) {
	row := Conn.DB.QueryRow(`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`, u.Name, u.Email)
	err := row.Scan(&u.ID)
	return u, err
}

func AddUserAddress(a UserAddress) error {
	_, err := Conn.DB.Exec(`INSERT INTO user_addresses (user_id, chain, address) VALUES ($1, $2, $3)`,
		a.UserID, a.Chain, a.Address)
	return err
}

// BALANCE
func GetUserBalances(userID string) ([]Balance, error) {
	query := `
	SELECT currency, SUM(CASE direction WHEN 'credit' THEN amount ELSE -amount END) as amount
	FROM ledger_entries
	JOIN transactions ON transactions.id = ledger_entries.transaction_id
	WHERE transactions.user_id = $1
	GROUP BY currency`
	rows, err := Conn.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}(rows)

	var balances []Balance
	for rows.Next() {
		var b Balance
		if err := rows.Scan(&b.Currency, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, nil
}

// TRANSACTIONS
func ProcessTransaction(tx TransactionRequest) error {
	txID := uuid.New().String()
	_, err := Conn.DB.Exec(`
		INSERT INTO transactions (id, user_id, type, amount, currency, status, tx_hash)
		VALUES ($1, $2, $3, $4, $5, 'completed', $6)`,
		txID, tx.UserID, tx.Type, tx.Amount, tx.Currency, tx.TxHash)
	if err != nil {
		return err
	}

	// Double-entry
	var credit, debit string
	switch tx.Type {
	case "deposit":
		credit = fmt.Sprintf("user:%d", tx.UserID)
		debit = "external"
	case "withdrawal":
		credit = "external"
		debit = fmt.Sprintf("user:%d", tx.UserID)
	default:
		return fmt.Errorf("unsupported type")
	}

	if err := insertLedgerEntry(txID, credit, tx.Currency, tx.Amount, "credit"); err != nil {
		return err
	}
	if err := insertLedgerEntry(txID, debit, tx.Currency, tx.Amount, "debit"); err != nil {
		return err
	}
	return nil
}

func ProcessTransfer(req TransferRequest) error {
	txID := uuid.New().String()
	_, err := Conn.DB.Exec(`INSERT INTO transactions (id, user_id, type, amount, currency, status)
	VALUES ($1, $2, 'transfer', $3, $4, 'completed')`, txID, req.FromUserID, req.Amount, req.Currency)
	if err != nil {
		return err
	}
	// debit sender
	from := fmt.Sprintf("user:%d", req.FromUserID)
	to := fmt.Sprintf("user:%d", req.ToUserID)
	if err := insertLedgerEntry(txID, from, req.Currency, req.Amount, "debit"); err != nil {
		return err
	}
	if err := insertLedgerEntry(txID, to, req.Currency, req.Amount, "credit"); err != nil {
		return err
	}
	return nil
}

func insertLedgerEntry(txID, account, currency string, amount float64, direction string) error {
	_, err := Conn.DB.Exec(`INSERT INTO ledger_entries (id, transaction_id, account, currency, amount, direction)
	VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), txID, account, currency, amount, direction)
	return err
}

func GetAddressTxs(address string) ([]map[string]interface{}, error) {
	rows, err := Conn.DB.Query(`
		SELECT tx_hash, amount, currency, direction, block_height, confirmed, created_at
		FROM onchain_transactions
		WHERE address = $1
		ORDER BY created_at DESC
	`, strings.ToLower(address))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var txHash, currency, direction string
		var amount float64
		var blockHeight int
		var confirmed bool
		var createdAt time.Time

		if err := rows.Scan(&txHash, &amount, &currency, &direction, &blockHeight, &confirmed, &createdAt); err != nil {
			return nil, err
		}
		tx := map[string]interface{}{
			"tx_hash":      txHash,
			"amount":       amount,
			"currency":     currency,
			"direction":    direction,
			"block_height": blockHeight,
			"confirmed":    confirmed,
			"timestamp":    createdAt,
		}
		results = append(results, tx)
	}
	return results, nil
}

func GetOnChainBalance(address string) (map[string]float64, error) {
	query := `
	SELECT currency, SUM(
		CASE direction WHEN 'credit' THEN amount ELSE -amount END
	) as balance
	FROM onchain_transactions
	WHERE address = $1 AND confirmed = true
	GROUP BY currency`
	rows, err := Conn.DB.Query(query, strings.ToLower(address))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make(map[string]float64)
	for rows.Next() {
		var currency string
		var balance float64
		if err := rows.Scan(&currency, &balance); err != nil {
			return nil, err
		}
		balances[currency] = balance
	}
	return balances, nil
}

func ReconcileAll() ([]map[string]string, error) {
	query := `
	SELECT o.address, o.tx_hash, o.currency, o.amount, o.direction
	FROM onchain_transactions o
	LEFT JOIN transactions t ON t.tx_hash = o.tx_hash
	WHERE t.tx_hash IS NULL AND o.confirmed = true
	`

	rows, err := Conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}(rows)

	var mismatches []map[string]string
	for rows.Next() {
		var address, txHash, currency, direction string
		var amount float64

		if err := rows.Scan(&address, &txHash, &currency, &amount, &direction); err != nil {
			return nil, err
		}
		mismatches = append(mismatches, map[string]string{
			"address":   address,
			"tx_hash":   txHash,
			"currency":  currency,
			"amount":    fmt.Sprintf("%.6f", amount),
			"direction": direction,
			"reason":    "unmatched on-chain transaction",
		})
	}
	return mismatches, nil
}
