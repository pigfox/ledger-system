package test

import (
	"github.com/stretchr/testify/assert"
	"ledger-system/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testReconciliation(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("POST", "/api/v1/reconciliation", nil)
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"Matched":`)
}
