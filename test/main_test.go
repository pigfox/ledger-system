package test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"ledger-system/internal/config"
	"ledger-system/internal/constants"
	"ledger-system/internal/db"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

var testDB *db.DB

var mainAddress = "0xdadB0d80178819F2319190D340ce9A924f783711"
var mainTxHash = "0xc17dce8502f989fde54da9922bc36a2767d0ae5b7ecf7904e49ff99aa19ad4e7"
var userID1 = 1 // Default user ID for tests
var userID2 = 2
var testCtx = context.Background()

func init() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println(".env file not found or could not be loaded")
	}

	config.CfgTest = config.LoadCfgTest()
	// Ensure the test DB is created before running tests
	createTestDB()
}

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env")

	// Load test config
	config.CfgTest = config.LoadCfgTest()
	config.Cfg = config.CfgTest

	// Connect to test DB
	testDB = db.ConnectTest()
	defer func() {
		if err := testDB.Conn.Close(); err != nil {
			log.Fatalf("Failed to close test DB connection: %v", err)
		}
		log.Println("Test DB connection closed")
	}()

	log.Printf("Loading schema from: %s", constants.InitSchema)

	schema, err := os.ReadFile("../" + constants.InitSchema)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	log.Println("Executing schema...")
	if _, err := testDB.Conn.Exec(string(schema)); err != nil {
		log.Fatalf("Failed to initialize test DB schema: %v", err)
	}
	log.Println("✅ Test DB schema initialized")

	// Clean up all tables
	tables := []string{"ledger_entries", "onchain_transactions", "transactions", "user_addresses", "users"}
	for _, table := range tables {
		_, err := testDB.Conn.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Printf("Failed to truncate table %s: %v", table, err)
		}
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func createTestDB() {
	var db *sql.DB
	var err error

	// Wait for DB to be ready
	maxAttempts := 10
	for i := 1; i <= maxAttempts; i++ {
		db, err = sql.Open("postgres", config.CfgTest.DBUrl)
		if err == nil && db.Ping() == nil {
			break
		}
		log.Printf("Waiting for test DB (%d/%d)...", i, maxAttempts)
		time.Sleep(2 * time.Second)
	}
	if err != nil || db.Ping() != nil {
		log.Fatalf("Test DB not ready after %d attempts", maxAttempts)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close test DB connection: %v", err)
		} else {
			log.Println("Test DB connection closed")
		}
	}(db)

	// Create DB if it doesn't exist
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE %s`, config.CfgTest.DBName))
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Fatalf("Failed to create test DB: %v", err)
	}

	// Now connect to test DB and drop all tables
	testDB, err := sql.Open("postgres", config.CfgTest.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}
	defer func(testDB *sql.DB) {
		err := testDB.Close()
		if err != nil {
			log.Fatalf("Failed to close test DB connection: %v", err)
		} else {
			log.Println("Test DB connection closed")
		}
	}(testDB)

	_, err = testDB.Exec(`
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		log.Fatalf("Failed to clean test DB: %v", err)
	}
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
	t.Run("GetUserBalances", testGetUserBalances)
	t.Run("GetUserBalanceByCurrency", testGetUserBalanceByCurrency)
	t.Run("GetAddressTransactions", testGetAddressTransactions)
	t.Run("GetAddressBalance", testGetAddressBalance)
	t.Run("Reconciliation", testReconciliation)
}

func truncateTables() {
	// Clean DB before test run
	_, err := testDB.Conn.Exec(`
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		log.Fatalf("Failed to clean test DB: %v", err)
	}
}
