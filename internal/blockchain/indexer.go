package blockchain

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Indexer structure
type Indexer struct {
	Client   *ethclient.Client
	DB       *sql.DB
	Interval time.Duration
}

// New creates a new indexer instance
func New(rpcURL string, db *sql.DB, interval time.Duration) *Indexer {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}
	return &Indexer{
		Client:   client,
		DB:       db,
		Interval: interval,
	}
}

// Start launches the indexer loop
func (i *Indexer) Start(ctx context.Context) {
	log.Println("Starting indexer loop")
	ticker := time.NewTicker(i.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			header, err := i.Client.HeaderByNumber(ctx, nil)
			if err != nil {
				log.Printf("Error fetching block header: %v", err)
				continue
			}
			if header == nil {
				continue
			}
			log.Printf("New Block: %v", header.Number)
			if err := i.processBlock(header.Number); err != nil {
				log.Printf("Block processing failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("Indexer stopped")
			return
		}
	}
}

// FetchMonitoredAddresses loads all addresses to track
func (i *Indexer) fetchMonitoredAddresses() (map[string]bool, error) {
	rows, err := i.DB.Query("SELECT LOWER(address) FROM user_addresses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make(map[string]bool)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		addresses[addr] = true
	}
	return addresses, nil
}

// Process a single block
func (i *Indexer) processBlock(blockNumber *big.Int) error {
	block, err := i.Client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return err
	}

	monitored, err := i.fetchMonitoredAddresses()
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions() {
		to := ""
		if tx.To() != nil {
			to = strings.ToLower(tx.To().Hex())
		}
		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			log.Printf("Failed to extract sender: %v", err)
			continue
		}
		fromHex := strings.ToLower(from.Hex())

		// Match to or from
		if monitored[to] || monitored[fromHex] {
			direction := "credit"
			address := to
			if monitored[fromHex] && !monitored[to] {
				direction = "debit"
				address = fromHex
			}

			log.Printf("Match TX %s | %s | %s %s",
				tx.Hash().Hex(), direction, tx.Value().String(), tx.To())

			// Insert into DB
			_, err := i.DB.Exec(`
				INSERT INTO onchain_transactions
				(id, address, tx_hash, amount, currency, direction, block_height, confirmed)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (tx_hash, address) DO NOTHING
			`, uuid.New(), address, tx.Hash().Hex(), weiToEther(tx.Value()), "ETH", direction, block.NumberU64(), true)

			if err != nil {
				log.Printf("DB insert failed: %v", err)
			}
		}
	}
	return nil
}

// Convert Wei (big.Int) to float64 ETH
func weiToEther(wei *big.Int) float64 {
	f, _ := new(big.Float).SetInt(wei).Float64()
	return f / 1e18
}
