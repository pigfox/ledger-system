package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"ledger-system/internal/config"
	"log"
)

var Conn Connection

type Connection struct {
	DB *sql.DB
}

func init() {
	db, err := sql.Open("postgres", config.Cfg.DBUrl)
	if err != nil {
		log.Fatal("DB connect error:", err)
	}
	Conn.DB = db
}

// InitIfNeeded checks if the schema exists and initializes it if not.
func InitIfNeeded(db *DB, schemaFilePath string) error {
	var exists bool
	err := db.Conn.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'users'
	)`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check schema existence: %w", err)
	}

	if exists {
		log.Println("✅ Schema already exists, skipping initialization")
		return nil
	}

	log.Printf("🔧 Loading schema from: %s", schemaFilePath)
	if err := db.InitSchema(schemaFilePath); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	log.Println("✅ Schema successfully initialized")
	return nil
}
