package test

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

const (
	dbName     = "xyzledger_test"
	dbUser     = "xyz"
	dbPassword = "xyz"
	dbPort     = "5432"
	dbHost     = "localhost"
)

func init() {
	_ = godotenv.Load("../.env")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	fmt.Println(dbPassword)
}

func TestMain(m *testing.M) {
	// Create test DB
	createTestDB()

	// Connect to it
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations (or load schema)
	schema, _ := os.ReadFile("migrations/001_init.sql")
	if _, err := testDB.Exec(string(schema)); err != nil {
		log.Fatalf("Failed to initialize test DB schema: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	err = testDB.Close()
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
		return
	}
	dropTestDB()

	os.Exit(code)
}

func createTestDB() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", dbUser, dbPassword, dbHost, dbPort)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Failed to close test DB", err)
		}
	}(db)

	// Terminate existing connections (if any)
	_, _ = db.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		  AND pid <> pg_backend_pid();`, dbName))

	_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER %s", dbName, dbUser)); err != nil {
		log.Fatalf("Failed to create test DB: %v", err)
	}
	time.Sleep(1 * time.Second) // allow time for DB to become available
}

func dropTestDB() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", dbUser, dbPassword, dbHost, dbPort)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Warning: Failed to reconnect to drop DB: %v", err)
		return
	}
	defer db.Close()

	_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
}
