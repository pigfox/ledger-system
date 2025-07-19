package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"ledger-system/internal/db"
	"net/http"
)

func GetUserBalances(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	balances, err := db.GetUserBalances(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(balances)
}
