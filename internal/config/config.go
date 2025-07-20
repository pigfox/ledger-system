package config

import (
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
	dbUrl := os.Getenv("DATABASE_URL")
	rpcUrl := os.Getenv("RPC_URL")
	apiKey := os.Getenv("API_KEY")
	if dbUrl == "" || rpcUrl == "" || apiKey == "" {
		log.Fatal("Required environment variables DATABASE_URL, RPC_URL, and API_KEY are not set")
	}

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
