package test

import (
	"bytes"
	"encoding/json"
	"ledger-system/internal/api"
	"ledger-system/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	api.RegisterRoutes(r, testDB)
	return r
}

func testCreateUsers(t *testing.T) {
	truncateTables()
	router := setupRouter()

	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}

	for _, user := range users {
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", config.CfgTest.APIKEY)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Errorf("CreateUser failed with status %d, body: %s", resp.Code, resp.Body.String())
		}
		assert.Equal(t, http.StatusCreated, resp.Code)
	}
}

func testAddUserAddresses(t *testing.T) {
	router := setupRouter()
	body := map[string]interface{}{
		"user_id": 1,
		"chain":   "ethereum",
		"address": mainAddress,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/addresses", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusCreated, resp.Code)
}

func testGetUserBalances(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/v1/users/1/balances", nil)
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"currency":"ETH"`)
}
