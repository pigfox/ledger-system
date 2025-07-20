package api

import (
	"github.com/gorilla/mux"
	"ledger-system/internal/constants"
	"net/http"
)

func RegisterRoutes(r *mux.Router) {
	v := constants.APIV1
	api := r.PathPrefix("/api/" + v).Subrouter()
	api.Use(APIKeyMiddleware)
	api.HandleFunc("/users", CreateUser).Methods("POST")
	api.HandleFunc("/transactions/deposit", Deposit).Methods("POST")
	api.HandleFunc("/transactions/withdraw", Withdraw).Methods("POST")
	api.HandleFunc("/transactions/transfer", Transfer).Methods("POST")
	api.HandleFunc("/users/{id}/balances", GetUserBalances).Methods("GET")

	api.HandleFunc("/addresses", RegisterAddress).Methods("POST")
	api.HandleFunc("/addresses/{address}/transactions", GetAddressTransactions).Methods("GET")
	api.HandleFunc("/addresses/{address}/balances", GetAddressBalance).Methods("GET")
	api.HandleFunc("/"+v+"/reconciliation", Reconcile).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
