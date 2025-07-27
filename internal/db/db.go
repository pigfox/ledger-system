// internal/db/db.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"ledger-system/internal/config"
	"ledger-system/internal/constants"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DB struct {
	Conn *sql.DB
}

func (d *DB) InitIfNeeded(schemaPath string) error {
	// Resolve absolute path to schema file relative to project root
	absPath, err := resolveSchemaPath(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to resolve schema path: %w", err)
	}

	log.Println("Loading schema from:", absPath)

	schema, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	log.Println("Executing schema...")
	if _, err := d.Conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	log.Println("Schema applied")
	return nil
}

func ConnectWithAutoCreate(curCfg config.Config) *DB {
	if curCfg.DBName == "" {
		log.Fatal("config.DBName is empty")
	}

	// Connect to admin DB
	adminURL := strings.Replace(curCfg.DBUrl, curCfg.DBName, "postgres", 1)
	adminDB, err := sql.Open("postgres", adminURL)
	if err != nil {
		log.Fatalf("Failed to connect to admin DB: %v", err)
	}
	defer adminDB.Close()

	// Check if DB exists
	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS(SELECT FROM pg_database WHERE datname = $1)", curCfg.DBName).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if DB exists: %v", err)
	}

	if !exists {
		log.Printf("DB '%s' does not exist. Creating...", curCfg.DBName)
		if _, err := adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", curCfg.DBName)); err != nil {
			log.Fatalf("Failed to create DB: %v", err)
		}
		log.Printf("DB '%s' created", curCfg.DBName)
	}

	// Connect to the actual DB
	db, err := sql.Open("postgres", curCfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Ping to DB failed: %v", err)
	}
	log.Println("Connected to DB successfully")

	wrapped := &DB{Conn: db}

	if err := wrapped.InitIfNeeded(constants.InitSchema); err != nil {
		log.Fatalf("DB schema initialization failed: %v", err)
	}

	return wrapped
}

func Connect() *DB {
	return ConnectWithAutoCreate(config.Cfg)
}

func ConnectTest() *DB {
	return ConnectWithAutoCreate(config.CfgTest)
}

func (d *DB) CreateUser(ctx context.Context, u User) (User, error) {
	row := d.Conn.QueryRowContext(ctx, `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`, u.Name, u.Email)
	err := row.Scan(&u.ID)
	return u, err
}

func (db *DB) FindUserByAddress(ctx context.Context, addr string) (int, error) {
	var userID int
	err := db.Conn.QueryRowContext(ctx, `
		SELECT user_id FROM user_addresses WHERE LOWER(address) = LOWER($1)
	`, addr).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (d *DB) AddUserAddress(ctx context.Context, a *UserAddress) error {
	return d.Conn.QueryRowContext(ctx, `
		INSERT INTO user_addresses (user_id, chain, address)
		VALUES ($1, $2, $3)
		RETURNING id`,
		a.UserID, a.Chain, a.Address).Scan(&a.ID)
}

func (d *DB) GetUserBalances(ctx context.Context, userID int, currency string) ([]Balance, error) {
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
		rows, err = d.Conn.QueryContext(ctx, query, account, currency)
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

func (d *DB) ProcessTransaction(ctx context.Context, tx TransactionRequest) (string, error) {
	txID := uuid.New().String()
	now := time.Now()

	txDB, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			txDB.Rollback()
		}
	}()

	_, err = txDB.ExecContext(ctx, `
		INSERT INTO transactions (id, user_id, type, amount, currency, status, tx_hash, block_height, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, txID, tx.UserID, tx.Type, tx.Amount, tx.Currency, "completed", tx.TxHash, tx.BlockHeight, now)
	if err != nil {
		return "", err
	}

	var creditAccount, debitAccount string
	switch tx.Type {
	case constants.Debit, constants.Deposit:
		creditAccount = fmt.Sprintf("user:%d", tx.UserID)
		debitAccount = constants.External
	case constants.Withdrawal:
		creditAccount = constants.External
		debitAccount = fmt.Sprintf("user:%d", tx.UserID)
	default:
		return "", fmt.Errorf("unsupported transaction type: %s", tx.Type)
	}

	creditEntry := &LedgerEntry{
		TransactionID: txID,
		Account:       creditAccount,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Direction:     constants.Credit,
		CreatedAt:     now,
	}
	if err := d.InsertLedgerEntryTx(ctx, txDB, creditEntry); err != nil {
		return "", err
	}

	debitEntry := &LedgerEntry{
		TransactionID: txID,
		Account:       debitAccount,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Direction:     constants.Debit,
		CreatedAt:     now,
	}
	if err := d.InsertLedgerEntryTx(ctx, txDB, debitEntry); err != nil {
		return "", err
	}

	if err := txDB.Commit(); err != nil {
		return "", err
	}

	return txID, nil
}

func (d *DB) InsertLedgerEntry(ctx context.Context, entry *LedgerEntry) error {
	entry.ID = uuid.New().String()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	_, err := d.Conn.ExecContext(ctx, `
		INSERT INTO ledger_entries (id, transaction_id, account, amount, currency, direction, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		entry.ID,
		entry.TransactionID,
		entry.Account,
		entry.Amount,
		entry.Currency,
		entry.Direction,
		entry.CreatedAt,
	)
	return err
}

func (d *DB) InsertLedgerEntryTx(ctx context.Context, tx *sql.Tx, entry *LedgerEntry) error {
	entry.ID = uuid.New().String()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ledger_entries (id, transaction_id, account, amount, currency, direction, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		entry.ID,
		entry.TransactionID,
		entry.Account,
		entry.Amount,
		entry.Currency,
		entry.Direction,
		entry.CreatedAt,
	)
	return err
}

func (d *DB) ProcessTransfer(ctx context.Context, req TransferRequest) (string, error) {
	txID := uuid.New().String()
	dbTx, err := d.Conn.Begin()
	if err != nil {
		return "", err
	}

	_, err = dbTx.ExecContext(ctx, `
		INSERT INTO transactions (id, user_id, type, amount, currency, status)
		VALUES ($1, $2, 'transfer', $3, $4, 'completed')`,
		txID, req.FromUserID, req.Amount, req.Currency)
	if err != nil {
		dbTx.Rollback()
		return "", err
	}

	now := time.Now()

	// Debit from sender
	entryFrom := &LedgerEntry{
		TransactionID: txID,
		Account:       fmt.Sprintf("user:%d", req.FromUserID),
		Amount:        req.Amount,
		Currency:      req.Currency,
		Direction:     "debit",
		CreatedAt:     now,
	}
	if err = d.InsertLedgerEntryTx(ctx, dbTx, entryFrom); err != nil {
		dbTx.Rollback()
		return "", err
	}

	// Credit to recipient
	entryTo := &LedgerEntry{
		TransactionID: txID,
		Account:       fmt.Sprintf("user:%d", req.ToUserID),
		Amount:        req.Amount,
		Currency:      req.Currency,
		Direction:     "credit",
		CreatedAt:     now,
	}
	if err = d.InsertLedgerEntryTx(ctx, dbTx, entryTo); err != nil {
		dbTx.Rollback()
		return "", err
	}

	if err = dbTx.Commit(); err != nil {
		return "", err
	}
	return txID, nil
}

func (d *DB) GetAddressTxs(ctx context.Context, address string) ([]map[string]interface{}, error) {
	rows, err := d.Conn.QueryContext(ctx, `
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

func (d *DB) GetOnChainBalance(ctx context.Context, address string) (map[string]float64, error) {
	query := `
	SELECT currency, SUM(
		CASE direction WHEN 'credit' THEN amount ELSE -amount END
	) as balance
	FROM onchain_transactions
	WHERE address = $1 AND confirmed = true
	GROUP BY currency`
	rows, err := d.Conn.QueryContext(ctx, query, strings.ToLower(address))
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

func (d *DB) TruncateAllTables() {
	tables := []string{"ledger_entries", "onchain_transactions", "transactions", "user_addresses", "users"}
	for _, table := range tables {
		if _, err := d.Conn.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			log.Printf("Failed to truncate %s: %v", table, err)
		}
	}
}

func (d *DB) Close() error {
	if d.Conn != nil {
		return d.Conn.Close()
	}
	return nil
}

func (d *DB) Drop() error {
	if d.Conn != nil {
		_, err := d.Conn.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			return fmt.Errorf("failed to drop schema: %w", err)
		}
		log.Println("Dropped and recreated public schema")
		return nil
	}
	return fmt.Errorf("no connection to drop schema")
}

func (d *DB) LoadSchema(schema string) {
	if d.Conn == nil {
		log.Fatal("no database connection")
	}

	_, err := d.Conn.Exec(schema)
	if err != nil {
		log.Fatalf("failed to execute schema: %v", err)
	}

	log.Println("Schema executed successfully")
}

func resolveSchemaPath(rel string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Search upward until we find the project root (where go.mod lives)
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory of %s", cwd)
		}
		dir = parent
	}

	abs := filepath.Join(dir, rel)
	return abs, nil
}
