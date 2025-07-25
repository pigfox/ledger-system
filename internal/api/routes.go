package api

import (
	"context"
	"github.com/gorilla/mux"
	"ledger-system/internal/constants"
	"ledger-system/internal/db"
	"net/http"
)

type Handler struct {
	DB  *db.DB
	Ctx context.Context
}

func NewHandler(ctx context.Context, db *db.DB) *Handler {
	return &Handler{
		DB:  db,
		Ctx: ctx,
	}
}

func RegisterRoutes(ctx context.Context, r *mux.Router, db *db.DB) http.Handler {
	v := constants.APIV1
	api := r.PathPrefix("/api/" + v).Subrouter()
	api.Use(RecoverMiddleware)
	api.Use(ContextTimeoutMiddleware(constants.TimeOut))
	api.Use(CheckAPIKeyMiddleware)
	h := NewHandler(ctx, db)
	api.HandleFunc("/users", h.CreateUser).Methods("POST")
	api.HandleFunc("/transactions/deposit", h.Deposit).Methods("POST")
	api.HandleFunc("/transactions/withdraw", h.Withdraw).Methods("POST")
	api.HandleFunc("/transactions/transfer", h.Transfer).Methods("POST")
	api.HandleFunc("/users/{id}/balances", h.GetUserBalances).Methods("GET")

	api.HandleFunc("/addresses", h.RegisterAddress).Methods("POST")
	api.HandleFunc("/addresses/{address}/transactions", h.GetAddressTransactions).Methods("GET")
	api.HandleFunc("/addresses/{address}/balances", h.GetAddressBalance).Methods("GET")
	api.HandleFunc("/reconciliation", h.ReconcileHandler).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	return r
}
