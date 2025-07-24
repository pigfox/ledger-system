// internal/db/db.go
package db

import (
	"database/sql"
	"fmt"
	"ledger-system/internal/config"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DB struct {
	Conn *sql.DB
}

// InitSchema loads the schema if required
func (d *DB) InitSchema(path string) error {
	log.Println("Loading schema from:", path) // ✅ Add this line

	schema, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	log.Println("Executing schema...")
	_, err = d.Conn.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	log.Println("Schema applied")
	return nil
}

func Connect() *DB {
	db, err := sql.Open("postgres", config.Cfg.DBUrl)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	return &DB{Conn: db}
}

func ConnectTest() *DB {
	db, err := sql.Open("postgres", config.CfgTest.DBUrl)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	return &DB{Conn: db}
}

func (d *DB) CreateUser(u User) (User, error) {
	row := d.Conn.QueryRow(`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`, u.Name, u.Email)
	err := row.Scan(&u.ID)
	return u, err
}

func (d *DB) AddUserAddress(a *UserAddress) error {
	return d.Conn.QueryRow(`
		INSERT INTO user_addresses (user_id, chain, address)
		VALUES ($1, $2, $3)
		RETURNING id`,
		a.UserID, a.Chain, a.Address).Scan(&a.ID)
}

func (d *DB) GetUserBalances(userID int, currency string) ([]Balance, error) {
	account := fmt.Sprintf("user:%d", userID)

	var rows *sql.Rows
	var err error

	if currency != "" {
		query := `
			SELECT currency,
				   SUM(CASE
					     WHEN direction = 'credit' THEN amount
					     WHEN direction = 'debit' THEN -amount
				   END) AS amount
			FROM ledger_entries
			WHERE account = $1 AND currency = $2
			GROUP BY currency;`
		rows, err = d.Conn.Query(query, account, currency)
	} else {
		query := `
			SELECT currency,
				   SUM(CASE
					     WHEN direction = 'credit' THEN amount
					     WHEN direction = 'debit' THEN -amount
				   END) AS amount
			FROM ledger_entries
			WHERE account = $1
			GROUP BY currency;`
		rows, err = d.Conn.Query(query, account)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []Balance
	for rows.Next() {
		var b Balance
		if err := rows.Scan(&b.Currency, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}

	if len(balances) == 0 && currency != "" {
		return []Balance{{Currency: currency, Amount: 0}}, nil
	}

	return balances, nil
}

func (d *DB) ProcessTransaction(tx TransactionRequest) (string, error) {
	txID := uuid.New().String()

	_, err := d.Conn.Exec(`
		INSERT INTO transactions (id, user_id, type, amount, currency, status, tx_hash)
		VALUES ($1, $2, $3, $4, $5, 'completed', $6)
	`, txID, tx.UserID, tx.Type, tx.Amount, tx.Currency, tx.TxHash)
	if err != nil {
		return "", err
	}

	var creditAccount, debitAccount string
	switch tx.Type {
	case "deposit":
		creditAccount = fmt.Sprintf("user:%d", tx.UserID)
		debitAccount = "external"
	case "withdrawal":
		creditAccount = "external"
		debitAccount = fmt.Sprintf("user:%d", tx.UserID)
	default:
		return "", fmt.Errorf("unsupported transaction type: %s", tx.Type)
	}

	creditEntry := &LedgerEntry{
		TransactionID: txID,
		Account:       creditAccount,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Direction:     "credit",
		CreatedAt:     time.Now(),
	}

	debitEntry := &LedgerEntry{
		TransactionID: txID,
		Account:       debitAccount,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Direction:     "debit",
		CreatedAt:     time.Now(),
	}

	if err := d.InsertLedgerEntry(creditEntry); err != nil {
		return "", err
	}
	if err := d.InsertLedgerEntry(debitEntry); err != nil {
		return "", err
	}

	return txID, nil
}

func (d *DB) InsertLedgerEntry(entry *LedgerEntry) error {
	_, err := d.Conn.Exec(`
		INSERT INTO ledger_entries (id, transaction_id, account, amount, currency, direction, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		uuid.New(),
		entry.TransactionID,
		entry.Account,
		entry.Amount,
		entry.Currency,
		entry.Direction,
		entry.CreatedAt,
	)
	return err
}

func (d *DB) ProcessTransfer(req TransferRequest) (string, error) {
	txID := uuid.New().String()
	dbTx, err := d.Conn.Begin()
	if err != nil {
		return "", err
	}

	_, err = dbTx.Exec(`
		INSERT INTO transactions (id, user_id, type, amount, currency, status)
		VALUES ($1, $2, 'transfer', $3, $4, 'completed')`,
		txID, req.FromUserID, req.Amount, req.Currency)
	if err != nil {
		dbTx.Rollback()
		return "", err
	}

	from := fmt.Sprintf("user:%d", req.FromUserID)
	to := fmt.Sprintf("user:%d", req.ToUserID)

	if err = insertLedgerEntryTx(dbTx, txID, from, req.Currency, req.Amount, "debit"); err != nil {
		dbTx.Rollback()
		return "", err
	}
	if err = insertLedgerEntryTx(dbTx, txID, to, req.Currency, req.Amount, "credit"); err != nil {
		dbTx.Rollback()
		return "", err
	}

	if err = dbTx.Commit(); err != nil {
		return "", err
	}
	return txID, nil
}

func (d *DB) GetAddressTxs(address string) ([]map[string]interface{}, error) {
	rows, err := d.Conn.Query(`
		SELECT tx_hash, amount, currency, direction, block_height, confirmed, created_at
		FROM onchain_transactions
		WHERE address = $1
		ORDER BY created_at DESC`, strings.ToLower(address))
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

func (d *DB) GetOnChainBalance(address string) (map[string]float64, error) {
	query := `
	SELECT currency, SUM(
		CASE direction WHEN 'credit' THEN amount ELSE -amount END
	) as balance
	FROM onchain_transactions
	WHERE address = $1 AND confirmed = true
	GROUP BY currency`
	rows, err := d.Conn.Query(query, strings.ToLower(address))
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("Failed to close rows:", err)
		}
	}(rows)

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
