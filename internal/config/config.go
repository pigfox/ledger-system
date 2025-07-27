package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"ledger-system/internal/constants"
	"log"
	"os"
	"path/filepath"
)

var Cfg Config
var CfgTest Config
var AllowedCurrencies map[string]bool

func init() {
	AllowedCurrencies = make(map[string]bool)
	AllowedCurrencies[constants.ETH] = true
	AllowedCurrencies[constants.MATIC] = true
	AllowedCurrencies[constants.USDC] = true
}

type Config struct {
	Port   string
	DBUrl  string
	DBName string
	RPCUrl string
	APIKEY string
}

func SetUp() {
	envPath := findEnvPath()

	if envPath == "" {
		log.Println(".env file not found in current or parent directories")
	} else {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Failed to load .env from %s: %v", envPath, err)
		} else {
			log.Println("Loaded .env from:", envPath)
		}
	}

	Cfg = LoadCfg()
	CfgTest = LoadCfgTest()
}

func LoadCfg() Config {
	pgUser := os.Getenv("POSTGRES_USER")
	pgPwd := os.Getenv("POSTGRES_PASSWORD")
	pgHost := os.Getenv("POSTGRES_HOST")
	pgDB := os.Getenv("POSTGRES_DB")
	pgPort := os.Getenv("POSTGRES_PORT")
	rpcUrl := os.Getenv("RPC_URL")
	apiKey := os.Getenv("API_KEY")
	if pgPort == "" || pgHost == "" || pgUser == "" || pgPwd == "" || pgDB == "" || rpcUrl == "" || apiKey == "" {
		log.Fatal("One or more required environment variables are missing (check POSTGRES_*, TEST_POSTGRES_*, RPC_URL, API_KEY)")
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
		DBName: pgDB,
	}
}

func LoadCfgTest() Config {
	pgUser := os.Getenv("TEST_POSTGRES_USER")
	pgPwd := os.Getenv("TEST_POSTGRES_PASSWORD")
	pgHost := os.Getenv("TEST_POSTGRES_HOST")
	pgDB := os.Getenv("TEST_POSTGRES_DB")
	pgPort := os.Getenv("TEST_POSTGRES_PORT")
	rpcUrl := os.Getenv("RPC_URL")
	apiKey := os.Getenv("API_KEY")
	if pgPort == "" || pgHost == "" || pgUser == "" || pgPwd == "" || pgDB == "" || rpcUrl == "" || apiKey == "" {
		log.Fatal("One or more required environment variables are missing (check POSTGRES_*, TEST_POSTGRES_*, RPC_URL, API_KEY)")
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
		DBName: pgDB,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func findEnvPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return ""
}
