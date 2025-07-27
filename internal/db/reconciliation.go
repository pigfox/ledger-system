package db

import (
	"context"
	"fmt"
	"ledger-system/internal/config"
	"ledger-system/internal/constants"
	"log"
	"time"
)

func (db *DB) ReconcileOnChainToLedger(ctx context.Context) (*ReconciliationReport, error) {
	rows, err := db.Conn.QueryContext(ctx, `
		SELECT id, address, tx_hash, amount, currency, direction, block_height
		FROM onchain_transactions
		WHERE reconciled IS FALSE
	`)
	if err != nil {
		return nil, fmt.Errorf("fetch unreconciled txs: %w", err)
	}
	defer rows.Close()

	report := &ReconciliationReport{
		Errors:       []string{},
		Incompatible: []OnChainTransaction{},
	}

	for rows.Next() {
		var tx OnChainTransaction
		if err := rows.Scan(&tx.ID, &tx.Address, &tx.TxHash, &tx.Amount, &tx.Currency, &tx.Direction, &tx.BlockHeight); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("scan error: %v", err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		userID, err := db.FindUserByAddress(ctx, tx.Address)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("unknown address: %s", tx.Address))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		if !isValidLedgerMatch(tx) {
			log.Printf("Incompatible TX: %s", tx.TxHash)
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		account := fmt.Sprintf("user:%d", userID)
		opposite := constants.Credit
		if tx.Direction == constants.Credit {
			opposite = constants.Debit
		}

		var exists int
		err = db.Conn.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM ledger_entries
			WHERE transaction_id = $1 AND account = $2
		`, tx.ID.String(), account).Scan(&exists)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("check existing ledger entry error: %v", err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}
		if exists > 0 {
			report.Matched++
			continue
		}

		txDB, err := db.Conn.BeginTx(ctx, nil)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("begin tx error: %v", err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		_, err = txDB.ExecContext(ctx, `
			INSERT INTO transactions (
				id, user_id, type, amount, currency, status, tx_hash, block_height, created_at
			)
			VALUES ($1, $2, 'reconciliation', $3, $4, 'completed', $5, $6, $7)
		`, tx.ID, userID, tx.Amount, tx.Currency, tx.TxHash, tx.BlockHeight, time.Now())
		if err != nil {
			txDB.Rollback()
			report.Errors = append(report.Errors, fmt.Sprintf("insert transaction error for %s: %v", tx.ID, err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		userEntry := &LedgerEntry{
			TransactionID: tx.ID.String(),
			Account:       account,
			Amount:        tx.Amount,
			Currency:      tx.Currency,
			Direction:     tx.Direction,
			CreatedAt:     time.Now(),
		}
		if err := db.InsertLedgerEntryTx(ctx, txDB, userEntry); err != nil {
			txDB.Rollback()
			report.Errors = append(report.Errors, fmt.Sprintf("ledger insert error (user) for %s: %v", tx.ID, err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		externalEntry := &LedgerEntry{
			TransactionID: tx.ID.String(),
			Account:       constants.External,
			Amount:        tx.Amount,
			Currency:      tx.Currency,
			Direction:     opposite,
			CreatedAt:     time.Now(),
		}
		if err := db.InsertLedgerEntryTx(ctx, txDB, externalEntry); err != nil {
			txDB.Rollback()
			report.Errors = append(report.Errors, fmt.Sprintf("ledger insert error (external) for %s: %v", tx.ID, err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		_, err = txDB.ExecContext(ctx, `
			UPDATE onchain_transactions SET reconciled = true WHERE id = $1
		`, tx.ID)
		if err != nil {
			txDB.Rollback()
			report.Errors = append(report.Errors, fmt.Sprintf("reconcile update error for %s: %v", tx.ID, err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		if err := txDB.Commit(); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("tx commit error for %s: %v", tx.ID, err))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		report.Matched++
	}

	report.Flagged = len(report.Incompatible)
	return report, nil
}

func isValidLedgerMatch(tx OnChainTransaction) bool {
	if tx.Amount <= 0 {
		log.Printf("Amount invalid: %f", tx.Amount)
		return false
	}

	if !config.IsValidCurrency(tx.Currency) {
		log.Printf("Currency invalid: %s", tx.Currency)
		return false
	}

	if tx.Direction != constants.Credit && tx.Direction != constants.Debit {
		log.Printf("Direction invalid: %s", tx.Direction)
		return false
	}

	if len(tx.TxHash) != 66 || tx.TxHash[:2] != "0x" {
		log.Printf("TxHash invalid format: %s", tx.TxHash)
		return false
	}

	//log.Printf("Valid TX: %s | %s | %f %s", tx.TxHash, tx.Direction, tx.Amount, tx.Currency)
	return true
}
