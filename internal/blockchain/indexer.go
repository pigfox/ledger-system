package blockchain

import (
	"context"
	"database/sql"
	"ledger-system/internal/constants"
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

	// Backfill last 1000 blocks
	head, err := i.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get latest block header: %v", err)
	}
	start := new(big.Int).Sub(head.Number, big.NewInt(constants.BackFillBlocksSize)) // last 1000 blocks

	for n := new(big.Int).Set(start); n.Cmp(head.Number) <= 0; n.Add(n, big.NewInt(1)) {
		if err := i.processBlock(n); err != nil {
			log.Printf("Backfill block %v failed: %v", n, err)
		}
	}

	log.Println("Backfill complete. Watching new blocks...")

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
		// Skip contract creation
		if tx.To() == nil {
			continue
		}

		to := strings.ToLower(tx.To().Hex())

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

		// Skip unmonitored txs
		if !monitored[to] && !monitored[fromHex] {
			continue
		}

		// Insert both directions if both are monitored
		if monitored[to] {
			log.Printf("Match TX %s | credit | %s ETH to %s", tx.Hash().Hex(), weiToEther(tx.Value()), to)
			res, err := i.DB.Exec(`
				INSERT INTO onchain_transactions
				(id, address, tx_hash, amount, currency, direction, block_height, confirmed)
				VALUES ($1, $2, $3, $4, $5, 'credit', $6, true)
				ON CONFLICT (tx_hash, address) DO NOTHING
			`, uuid.New(), to, tx.Hash().Hex(), weiToEther(tx.Value()), "ETH", block.NumberU64())
			if err != nil {
				log.Printf("Insert credit failed for tx %s: %v", tx.Hash().Hex(), err)
			} else {
				n, _ := res.RowsAffected()
				log.Printf("Inserted credit rows: %d", n)
			}
		}

		if monitored[fromHex] {
			log.Printf("Match TX %s | debit | %s ETH from %s", tx.Hash().Hex(), weiToEther(tx.Value()), fromHex)
			res, err := i.DB.Exec(`
				INSERT INTO onchain_transactions
				(id, address, tx_hash, amount, currency, direction, block_height, confirmed)
				VALUES ($1, $2, $3, $4, $5, 'debit', $6, true)
				ON CONFLICT (tx_hash, address) DO NOTHING
			`, uuid.New(), fromHex, tx.Hash().Hex(), weiToEther(tx.Value()), "ETH", block.NumberU64())
			if err != nil {
				log.Printf("Insert debit failed for tx %s: %v", tx.Hash().Hex(), err)
			} else {
				n, _ := res.RowsAffected()
				log.Printf("Inserted debit rows: %d", n)
			}
		}
	}

	return nil
}

func weiToEther(wei *big.Int) float64 {
	f, _ := new(big.Float).SetInt(wei).Float64()
	return f / 1e18
}
