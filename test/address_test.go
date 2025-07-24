package test

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"ledger-system/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testGetAddressTransactions(t *testing.T) {
	router := setupRouter()
	address := mainAddress
	req := httptest.NewRequest("GET", "/api/v1/addresses/"+address+"/transactions", nil)
	req = mux.SetURLVars(req, map[string]string{"address": address})
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), mainTxHash)
}

func testGetAddressBalance(t *testing.T) {
	router := setupRouter()
	address := mainAddress
	req := httptest.NewRequest("GET", "/api/v1/addresses/"+address+"/balances", nil)
	req = mux.SetURLVars(req, map[string]string{"address": address})
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"ETH":`)
}
