package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledger-system/internal/api"
	"ledger-system/internal/db"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *mux.Router {
	db.Connect()
	r := mux.NewRouter()
	api.RegisterRoutes(r)
	return r
}

func TestHealthEndpoint(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestCreateUserAPI(t *testing.T) {
	router := setupRouter()

	body := map[string]string{
		"name":  "Test User",
		"email": "testuser@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusCreated, resp.Code)
}
