package db

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (db *DB) ReconcileOnChainToLedger(ctx context.Context) (*ReconciliationReport, error) {
	rows, err := db.Conn.QueryContext(ctx, `
		SELECT id, address, tx_hash, amount, currency, direction
		FROM onchain_transactions
		WHERE reconciled = false
	`)
	if err != nil {
		return nil, fmt.Errorf("fetch unreconciled txs: %w", err)
	}
	defer rows.Close()

	report := &ReconciliationReport{}
	for rows.Next() {
		var tx OnChainTransaction
		if err := rows.Scan(&tx.ID, &tx.Address, &tx.TxHash, &tx.Amount, &tx.Currency, &tx.Direction); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		userID, err := db.FindUserByAddress(tx.Address)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("unknown address: %s", tx.Address))
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		entry := &LedgerEntry{
			TransactionID: tx.TxHash, // using tx_hash from onchain data
			Account:       fmt.Sprintf("user:%d", userID),
			Amount:        tx.Amount,
			Currency:      tx.Currency,
			Direction:     tx.Direction,
			CreatedAt:     time.Now(),
		}

		if err := db.InsertLedgerEntry(entry); err != nil {
			report.Errors = append(report.Errors, err.Error())
			report.Incompatible = append(report.Incompatible, tx)
			continue
		}

		if _, err := db.Conn.ExecContext(ctx, `
			UPDATE onchain_transactions SET reconciled = true WHERE id = $1
		`, tx.ID); err != nil {
			log.Printf("Failed to mark tx %s as reconciled: %v", tx.TxHash, err)
		}

		report.Matched++
	}

	report.Flagged = len(report.Incompatible)
	return report, nil
}
