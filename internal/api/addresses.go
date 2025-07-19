package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"ledger-system/internal/db"
	"net/http"
)

func RegisterAddress(w http.ResponseWriter, r *http.Request) {
	var req db.UserAddress
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid", 400)
		return
	}
	if err := db.AddUserAddress(req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetAddressTransactions(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]
	txs, err := db.GetAddressTxs(addr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.NewEncoder(w).Encode(txs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func GetAddressBalance(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]
	balance, err := db.GetOnChainBalance(addr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
