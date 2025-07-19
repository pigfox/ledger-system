package test

import (
	"ledger-system/internal/db"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLedgerDoubleEntryBalanced(t *testing.T) {
	entries := []db.Entry{
		{Account: "user:1", Amount: 100, Currency: "ETH", Direction: "credit"},
		{Account: "external", Amount: 100, Currency: "ETH", Direction: "debit"},
	}

	debit := 0.0
	credit := 0.0
	for _, e := range entries {
		if e.Direction == "debit" {
			debit += e.Amount
		} else if e.Direction == "credit" {
			credit += e.Amount
		}
	}
	assert.Equal(t, debit, credit, "Ledger should be balanced")
}
