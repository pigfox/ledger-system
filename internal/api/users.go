package api

import (
	"encoding/json"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	var req Req
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// Call db.CreateUser(...) and return result
	w.WriteHeader(http.StatusCreated)
}
