package api

import (
	"github.com/gorilla/mux"
	"ledger-system/internal/constants"
	"ledger-system/internal/db"
	"net/http"
)

type Handler struct {
	DB *db.DB
}

func NewHandler(database *db.DB) *Handler {
	return &Handler{DB: database}
}

func RegisterRoutes(r *mux.Router, db *db.DB) {
	v := constants.APIV1
	api := r.PathPrefix("/api/" + v).Subrouter()
	api.Use(APIKeyMiddleware)
	h := NewHandler(db)
	api.HandleFunc("/users", h.CreateUser).Methods("POST")
	api.HandleFunc("/transactions/deposit", h.Deposit).Methods("POST")
	api.HandleFunc("/transactions/withdraw", h.Withdraw).Methods("POST")
	api.HandleFunc("/transactions/transfer", h.Transfer).Methods("POST")
	api.HandleFunc("/users/{id}/balances", h.GetUserBalances).Methods("GET")

	api.HandleFunc("/addresses", h.RegisterAddress).Methods("POST")
	api.HandleFunc("/addresses/{address}/transactions", h.GetAddressTransactions).Methods("GET")
	api.HandleFunc("/addresses/{address}/balances", h.GetAddressBalance).Methods("GET")
	api.HandleFunc("/reconciliation", Reconcile).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")
}
