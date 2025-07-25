package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"ledger-system/internal/api"
	"ledger-system/internal/blockchain"
	"ledger-system/internal/config"
	"ledger-system/internal/constants"
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

	config.Cfg = config.LoadCfg()
	dbPg := db.Connect()
	router := mux.NewRouter()
	handler := api.RegisterRoutes(ctx, router, dbPg)

	if err := db.InitIfNeeded(dbPg, constants.InitSchema); err != nil {
		log.Fatalf("❌ DB schema initialization failed: %v", err)
	}

	indexer := blockchain.New(ctx, config.Cfg.RPCUrl, dbPg.Conn, 15*time.Second)
	go indexer.Start(ctx)

	log.Printf("Listening on :%s", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(":"+config.Cfg.Port, handler))
}
