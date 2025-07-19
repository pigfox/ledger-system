package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"ledger-system/internal/api"
	"ledger-system/internal/blockchain"
	"ledger-system/internal/config"
	"ledger-system/internal/db"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	_ = godotenv.Load(".env")
	cfg := config.Load()
	r := mux.NewRouter()

	api.RegisterRoutes(r)
	handler := api.RecoverMiddleware(r)
	sqlDB := db.Connect()
	indexer := blockchain.New(os.Getenv("RPC_URL"), sqlDB.DB, 15*time.Second)
	go indexer.Start(context.Background())

	log.Printf("✅ Listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
