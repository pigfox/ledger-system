package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) ReconcileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	report, err := h.DB.ReconcileOnChainToLedger(ctx)
	if err != nil {
		log.Printf("reconciliation failed: %v", err)
		http.Error(w, "internal reconciliation error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(report)
	if err != nil {
		log.Printf("failed to encode reconciliation report: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
