package api

import (
	"net/http"
)

func Reconcile(w http.ResponseWriter, r *http.Request) {
	/*
		report, err := db.ReconcileAll()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = json.NewEncoder(w).Encode(report)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	*/
}
