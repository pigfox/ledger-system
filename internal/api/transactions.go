package api

import (
	"encoding/json"
	"ledger-system/internal/db"
	"net/http"
)

func Deposit(w http.ResponseWriter, r *http.Request) {
	var tx db.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "invalid deposit", 400)
		return
	}
	tx.Type = "deposit"
	if err := db.ProcessTransaction(tx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	var tx db.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "invalid withdrawal", 400)
		return
	}
	tx.Type = "withdrawal"
	if err := db.ProcessTransaction(tx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	var tx db.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "invalid transfer", 400)
		return
	}
	if err := db.ProcessTransfer(tx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
