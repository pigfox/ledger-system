package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// IsSmartContract returns true if the address contains bytecode
func IsSmartContract(client *ethclient.Client, address common.Address) (bool, error) {
	code, err := client.CodeAt(context.Background(), address, nil) // latest block
	if err != nil {
		return false, fmt.Errorf("code lookup failed: %w", err)
	}
	return len(code) > 0, nil
}

// GetLatestBlockNumber returns the current latest block number
func GetLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

// WeiToEther converts Wei to Ether (float64)
func WeiToEther(wei *big.Int) float64 {
	eth := new(big.Float).SetInt(wei)
	divisor := big.NewFloat(1e18)
	result, _ := new(big.Float).Quo(eth, divisor).Float64()
	return result
}

// MustConnect returns a connected Ethereum client or logs fatal
func MustConnect(rpcURL string) *ethclient.Client {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", rpcURL, err)
	}
	return client
}
