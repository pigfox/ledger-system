package test

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"ledger-system/internal/config"
	"ledger-system/internal/constants"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testDepositFunds(t *testing.T) {
	router := setupRouter()

	body := map[string]interface{}{
		"user_id":  userID1,
		"amount":   100.0,
		"currency": constants.ETH,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("❌ Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/transactions/deposit", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.CfgTest.APIKEY)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	//t.Logf("DepositFunds response code: %d", resp.Code)
	//t.Logf("DepositFunds response body: %s", resp.Body.String())

	assert.Equal(t, http.StatusCreated, resp.Code, "Expected status 201 Created")
}

func testWithdrawFunds(t *testing.T) {
	router := setupRouter()
	body := map[string]interface{}{
		"user_id":  userID1,
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
		"from_user_id": userID1,
		"to_user_id":   userID2,
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

func seedTransaction(t *testing.T) {
	txID := uuid.New()

	_, err := testDB.Conn.ExecContext(testCtx, `
        INSERT INTO onchain_transactions (
            id, address, tx_hash, amount, currency, direction, block_height, reconciled, created_at
        ) VALUES (
            $1, $2, $3, $4, $5, 'credit', $6, false, NOW()
        )
    `, txID, mainAddress, mainTxHash, 100.0, "ETH", 12345678)
	if err != nil {
		t.Fatalf("Failed to seed onchain transaction: %v", err)
	}
}
