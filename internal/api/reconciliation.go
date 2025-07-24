package api

import (
	"context"
	"encoding/json"
	"net/http"
)

func (h *Handler) ReconcileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	report, err := h.DB.ReconcileOnChainToLedger(ctx)
	if err != nil {
		http.Error(w, "Reconciliation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}
