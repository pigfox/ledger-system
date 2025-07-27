package test

import (
	"context"
	_ "github.com/lib/pq"
	"ledger-system/internal/config"
	"ledger-system/internal/db"
	"log"
	"os"
	"strings"
	"testing"
)

var testDB *db.DB

var mainAddress = strings.ToLower("0xdadB0d80178819F2319190D340ce9A924f783711")
var mainTxHash = strings.ToLower("0xc17dce8502f989fde54da9922bc36a2767d0ae5b7ecf7904e49ff99aa19ad4e7")
var userID1 = 1 // Default user ID for tests
var userID2 = 2
var testCtx = context.Background()

func setUp() {
	config.CfgTest = config.LoadCfgTest()
	testDB = db.ConnectTest()
	testDB.TruncateAllTables()
}

func tearDown() {
	err := testDB.Drop()
	if err != nil {
		log.Printf("Failed to drop test DB: %v", err)
	}
}

func TestMain(m *testing.M) {
	config.SetUp()
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestEndToEndLedgerFlow(t *testing.T) {
	t.Run("Health", testHealth)
	t.Run("CreateUsers", testCreateUsers)
	t.Run("SeedTransaction", func(t *testing.T) {
		seedTransaction(t)
	})
	t.Run("AddUserAddresses", testAddUserAddresses)
	t.Run("DepositFunds", testDepositFunds)
	t.Run("WithdrawFunds", testWithdrawFunds)
	t.Run("TransferFunds", testTransferFunds)
	// Reconciliation must happen before reading balances/txs
	t.Run("Reconciliation", testReconciliation)
	t.Run("GetUserBalances", testGetUserBalances)
	t.Run("GetUserBalanceByCurrency", testGetUserBalanceByCurrency)
	t.Run("GetAddressTransactions", testGetAddressTransactions)
	t.Run("GetAddressBalance", testGetAddressBalance)
}
