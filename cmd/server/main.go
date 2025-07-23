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
	router := mux.NewRouter()
	dbPg := db.Connect()
	api.RegisterRoutes(router, dbPg)
	handler := api.RecoverMiddleware(router)

	var exists bool
	err := dbPg.Conn.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'users'
	)`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check schema existence: %v", err)
	}

	if !exists {
		log.Printf("Loading schema from: %s", constants.InitSchema)
		if err := dbPg.InitSchema(constants.InitSchema); err != nil {
			log.Fatalf("Failed to initialize schema: %v", err)
		}
		log.Println("Schema applied")
	} else {
		log.Println("Schema already exists, skipping initialization")
	}
	indexer := blockchain.New(ctx, config.Cfg.RPCUrl, dbPg.Conn, 15*time.Second)
	go indexer.Start(ctx)

	log.Printf("Listening on :%s", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(":"+config.Cfg.Port, handler))
}
