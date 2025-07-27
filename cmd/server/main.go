package main

import (
	"context"
	"github.com/gorilla/mux"
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
	config.SetUp()
	dbPg := db.Connect()
	router := mux.NewRouter()
	handler := api.RegisterRoutes(ctx, router, dbPg)

	indexer := blockchain.New(ctx, config.Cfg.RPCUrl, dbPg.Conn, 15*time.Second)
	go indexer.Start(ctx)

	log.Printf("Listening on :%s", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(":"+config.Cfg.Port, handler))
}
