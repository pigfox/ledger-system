package test

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"ledger-system/internal/config"
	"ledger-system/internal/db"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testDepositFunds(t *testing.T) {
	router := setupRouter()
	body := map[string]interface{}{
		"user_id":  userID1,
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
	tx := db.TransactionRequest{
		UserID:      userID1,
		Type:        "deposit",
		Amount:      100.0,
		Currency:    "ETH",
		TxHash:      mainTxHash,
		BlockHeight: 12345678,
	}

	_, err := testDB.ProcessTransaction(testCtx, tx)
	if err != nil {
		t.Fatalf("Failed to seed transaction: %v", err)
	}

	_, err = testDB.Conn.Exec(`
		INSERT INTO onchain_transactions (id, address, tx_hash, amount, currency, direction, block_height, confirmed, created_at)
		VALUES ($1, $2, $3, $4, $5, 'credit', 12345678, true, NOW())
	`, uuid.New(), strings.ToLower(mainAddress), mainTxHash, 100.0, "ETH")
	if err != nil {
		t.Fatalf("Failed to seed onchain transaction: %v", err)
	}
}
