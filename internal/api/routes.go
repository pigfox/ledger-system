package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users", CreateUser).Methods("POST")
	r.HandleFunc("/transactions/deposit", Deposit).Methods("POST")
	r.HandleFunc("/transactions/withdraw", Withdraw).Methods("POST")
	r.HandleFunc("/transactions/transfer", Transfer).Methods("POST")
	r.HandleFunc("/users/{id}/balances", GetUserBalances).Methods("GET")

	r.HandleFunc("/addresses", RegisterAddress).Methods("POST")
	r.HandleFunc("/addresses/{address}/transactions", GetAddressTransactions).Methods("GET")
	r.HandleFunc("/addresses/{address}/balance", GetAddressBalance).Methods("GET")

	r.HandleFunc("/reconciliation", Reconcile).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}
