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
	address := "0xabc123abc123abc123abc123abc123abc123abc1"
	req := httptest.NewRequest("GET", "/api/v1/addresses/"+address+"/transactions", nil)
	req = mux.SetURLVars(req, map[string]string{"address": address})
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"tx_hash"`)
}

func testGetAddressBalance(t *testing.T) {
	router := setupRouter()
	address := "0xdadB0d80178819F2319190D340ce9A924f783711"
	req := httptest.NewRequest("GET", "/api/v1/addresses/"+address+"/balances", nil)
	req = mux.SetURLVars(req, map[string]string{"address": address})
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"ETH":`)
}
