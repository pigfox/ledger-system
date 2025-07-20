package scripts

import (
	"encoding/json"
	"fmt"
	"log"

	"ledger-system/internal/db"
)

// Run performs a one-time reconciliation job using db.Conn.DB.
func Run() {
	if db.Conn.DB == nil {
		log.Fatal("DB not initialized — call db.Connect() before scripts.Run()")
	}

	fmt.Println("Running reconciliation...")
	report, err := db.ReconcileAll()
	if err != nil {
		log.Fatalf("Reconciliation failed: %v", err)
	}

	if len(report) == 0 {
		fmt.Println("All on-chain transactions are accounted for.")
		return
	}

	fmt.Printf("Found %d unmatched on-chain transactions:\n", len(report))
	out, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(out))
}
