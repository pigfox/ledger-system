package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"ledger-system/internal/api"
	"ledger-system/internal/blockchain"
	"ledger-system/internal/config"
	"ledger-system/internal/db"
	"ledger-system/recoverx"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()
	defer recoverx.RecoverAndLog(ctx)
	_ = godotenv.Load(".env")

	config.Cfg = config.Load()
	router := mux.NewRouter()

	api.RegisterRoutes(router)
	handler := api.RecoverMiddleware(router)
	sqlDB := db.Connect()
	indexer := blockchain.New(ctx, config.Cfg.RPCUrl, sqlDB.DB, 15*time.Second)
	go indexer.Start(ctx)

	log.Printf("Listening on :%s", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(":"+config.Cfg.Port, handler))
}
