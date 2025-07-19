package config

import "os"

type Config struct {
	Port   string
	DBUrl  string
	RPCUrl string
}

func Load() Config {
	return Config{
		Port:   getEnv("PORT", "8080"),
		DBUrl:  os.Getenv("DATABASE_URL"),
		RPCUrl: os.Getenv("RPC_URL"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
