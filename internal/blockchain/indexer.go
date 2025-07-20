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

type Indexer struct {
	Client   *ethclient.Client
	DB       *sql.DB
	Interval time.Duration
	ChainID  *big.Int
}

func New(ctx context.Context, rpcURL string, db *sql.DB, interval time.Duration) *Indexer {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch chain ID: %v", err)
	}

	return &Indexer{
		Client:   client,
		DB:       db,
		Interval: interval,
		ChainID:  chainID,
	}
}

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

func (i *Indexer) fetchMonitoredAddresses() (map[string]bool, error) {
	rows, err := i.DB.Query("SELECT LOWER(address) FROM user_addresses")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}(rows)

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

		if tx.ChainId() == nil {
			log.Printf("Skipping tx %s: missing chain ID", tx.Hash().Hex())
			continue
		}

		from, err := types.Sender(types.LatestSignerForChainID(i.ChainID), tx)
		if err != nil {
			log.Printf("Failed to extract sender for tx %s: %v", tx.Hash().Hex(), err)
			continue
		}
		fromHex := strings.ToLower(from.Hex())

		if monitored[to] || monitored[fromHex] {
			direction := "credit"
			address := to
			if monitored[fromHex] && !monitored[to] {
				direction = "debit"
				address = fromHex
			}

			log.Printf("Match TX %s | %s | %s %s",
				tx.Hash().Hex(), direction, tx.Value().String(), tx.To())

			_, err := i.DB.Exec(`
				INSERT INTO onchain_transactions
				(id, address, tx_hash, amount, currency, direction, block_height, confirmed)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (tx_hash, address) DO NOTHING
			`, uuid.New(), address, tx.Hash().Hex(), weiToEther(tx.Value()), "ETH", direction, block.NumberU64(), true)

			if err != nil {
				log.Printf("DB insert failed for tx %s: %v", tx.Hash().Hex(), err)
			}
		}
	}

	return nil
}

func weiToEther(wei *big.Int) float64 {
	f, _ := new(big.Float).SetInt(wei).Float64()
	return f / 1e18
}
