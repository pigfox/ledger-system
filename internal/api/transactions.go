package api

import (
	"encoding/json"
	"ledger-system/internal/db"
	"net/http"
)

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	var tx db.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid deposit request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if tx.UserID == 0 || tx.Amount <= 0 || tx.Currency == "" {
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	tx.Type = "deposit"
	id, err := h.DB.ProcessTransaction(tx)
	if err != nil {
		http.Error(w, "Failed to process transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var tx db.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid withdrawal request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if tx.UserID == 0 || tx.Amount <= 0 || tx.Currency == "" {
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	// Check balance
	balance, err := h.DB.GetUserBalances(tx.UserID, tx.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var available float64
	for _, b := range balance {
		if b.Currency == tx.Currency {
			available = b.Amount
			break
		}
	}

	if tx.Amount > available {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	tx.Type = "withdrawal"
	txID, err := h.DB.ProcessTransaction(tx)
	if err != nil {
		http.Error(w, "Failed to process transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"id": txID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var tx db.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid transfer request", http.StatusBadRequest)
		return
	}

	if tx.FromUserID == 0 || tx.ToUserID == 0 || tx.Amount <= 0 || tx.Currency == "" {
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	if tx.FromUserID == tx.ToUserID {
		http.Error(w, "Sender and recipient cannot be the same", http.StatusBadRequest)
		return
	}

	balance, err := h.DB.GetUserBalances(tx.FromUserID, tx.Currency)
	if err != nil {
		http.Error(w, "Failed to check balance: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var available float64
	for _, b := range balance {
		if b.Currency == tx.Currency {
			available = b.Amount
			break
		}

	}
	if tx.Amount > available {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// ✅ Process transfer
	txID, err := h.DB.ProcessTransfer(tx)
	if err != nil {
		http.Error(w, "Failed to process transfer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Respond with tx ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"id": txID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
