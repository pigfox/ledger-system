package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"ledger-system/internal/db"
	"net/http"
)

func (h *Handler) RegisterAddress(w http.ResponseWriter, r *http.Request) {
	var req db.UserAddress
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 || req.Chain == "" || req.Address == "" {
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	if err := h.DB.AddUserAddress(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(req); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) GetAddressTransactions(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]
	txs, err := h.DB.GetAddressTxs(addr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(txs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Handler) GetAddressBalance(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]
	balance, err := h.DB.GetOnChainBalance(addr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
