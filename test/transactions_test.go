package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"ledger-system/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testDepositFunds(t *testing.T) {
	router := setupRouter()
	body := map[string]interface{}{
		"user_id":  1,
		"amount":   100.0,
		"currency": "ETH",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/transactions/deposit", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusCreated, resp.Code)
}

func testWithdrawFunds(t *testing.T) {
	router := setupRouter()
	body := map[string]interface{}{
		"user_id":  1,
		"amount":   50.0,
		"currency": "ETH",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/transactions/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func testTransferFunds(t *testing.T) {
	router := setupRouter()
	body := map[string]interface{}{
		"from_user_id": 1,
		"to_user_id":   2,
		"amount":       25.0,
		"currency":     "ETH",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/transactions/transfer", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
