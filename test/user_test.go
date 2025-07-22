package test

import (
	"bytes"
	"encoding/json"
	"ledger-system/internal/api"
	"ledger-system/internal/config"
	"ledger-system/internal/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *mux.Router {
	db.Connect() // Optional if mocking everything
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

func TestCreateUser(t *testing.T) {
	router := setupRouter()

	body := map[string]string{
		"name":  "Test User",
		"email": "testuser@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.Cfg.APIKEY)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusCreated, resp.Code)
}

func TestAddUserAddress(t *testing.T) {
	// Make sure user with ID 1 exists first (insert manually in DB or in a setup function)

	body := map[string]interface{}{
		"user_id": 1,
		"chain":   "ethereum",
		"address": "0xabc123abc123abc123abc123abc123abc123abc1",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/addresses", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.Cfg.APIKEY)

	resp := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Contains(t, resp.Body.String(), `"user_id":1`)
}

func TestGetUserBalances(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/v1/users/2/balances", nil)
	req.Header.Set("X-API-Key", config.Cfg.APIKEY)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"currency":"ETH"`)
}
