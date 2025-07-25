package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"ledger-system/internal/db"
	"net/http"
	"strconv"
)

func (h *Handler) GetUserBalances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Optional query param ?currency=ETH
	currency := r.URL.Query().Get("currency")

	balances, err := h.DB.GetUserBalances(ctx, id, currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter out zero balances
	nonZero := make([]db.Balance, 0)
	for _, b := range balances {
		if b.Amount != 0 {
			nonZero = append(nonZero, b)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nonZero); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
