package config

import (
	"fmt"
	"log"
	"os"
)

var Cfg Config

type Config struct {
	Port   string
	DBUrl  string
	RPCUrl string
	APIKEY string
}

func Load() Config {
	pgUser := os.Getenv("POSTGRES_USER")
	pgPwd := os.Getenv("POSTGRES_PASSWORD")
	pgHost := os.Getenv("POSTGRES_HOST")
	pgDB := os.Getenv("POSTGRES_DB")
	pgPort := os.Getenv("POSTGRES_PORT")
	rpcUrl := os.Getenv("RPC_URL")
	apiKey := os.Getenv("API_KEY")
	if pgPort == "" || pgHost == "" || pgUser == "" || pgPwd == "" || pgDB == "" || rpcUrl == "" || apiKey == "" {
		log.Fatal("Required environment variables DATABASE_URL, RPC_URL, and API_KEY are not set")
	}
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgUser,
		pgPwd,
		pgHost,
		pgPort,
		pgDB,
	)

	return Config{
		Port:   getEnv("PORT", "8080"),
		DBUrl:  dbUrl,
		RPCUrl: rpcUrl,
		APIKEY: apiKey,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
