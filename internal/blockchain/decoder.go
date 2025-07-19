package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC-20 Transfer event signature: keccak256("Transfer(address,address,uint256)")
var transferEventSigHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55aeb4cdb48")

// Track known token contracts (USDC on Ethereum and Polygon)
var knownERC20Tokens = map[string]string{
	// Ethereum USDC
	"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": "USDC",
	// Polygon USDC
	"0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174": "USDC",
}

// ParseERC20Transfers parses all logs in a block and extracts relevant ERC-20 transfers
func ParseERC20Transfers(client *ethclient.Client, blockNumber *big.Int, monitored map[string]bool) ([]DecodedTransfer, error) {
	logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: blockNumber,
		ToBlock:   blockNumber,
		Topics:    [][]common.Hash{{transferEventSigHash}},
	})
	if err != nil {
		return nil, fmt.Errorf("filter logs error: %w", err)
	}

	var transfers []DecodedTransfer

	for _, vLog := range logs {
		if len(vLog.Topics) < 3 || len(vLog.Data) != 32 {
			continue // malformed log
		}

		// Extract addresses
		from := common.HexToAddress(vLog.Topics[1].Hex())
		to := common.HexToAddress(vLog.Topics[2].Hex())
		amount := new(big.Int).SetBytes(vLog.Data)

		fromStr := strings.ToLower(from.Hex())
		toStr := strings.ToLower(to.Hex())
		tokenStr := strings.ToLower(vLog.Address.Hex())

		// Only handle known tokens (like USDC)
		symbol, ok := knownERC20Tokens[tokenStr]
		if !ok {
			continue
		}

		// Check if one of the addresses is monitored
		if !monitored[fromStr] && !monitored[toStr] {
			continue
		}

		// Determine direction
		var direction, userAddress string
		if monitored[toStr] {
			direction = "credit"
			userAddress = toStr
		} else if monitored[fromStr] {
			direction = "debit"
			userAddress = fromStr
		}

		dec := DecodedTransfer{
			TokenAddress: tokenStr,
			UserAddress:  userAddress,
			TxHash:       vLog.TxHash.Hex(),
			BlockHeight:  vLog.BlockNumber,
			Amount:       normalizeUSDC(amount),
			Currency:     symbol,
			Direction:    direction,
		}
		transfers = append(transfers, dec)
	}

	return transfers, nil
}

// DecodedTransfer is a structured ERC-20 Transfer
type DecodedTransfer struct {
	TokenAddress string
	UserAddress  string
	TxHash       string
	BlockHeight  uint64
	Amount       float64
	Currency     string
	Direction    string // credit or debit
}

// USDC uses 6 decimals (not 18)
func normalizeUSDC(amount *big.Int) float64 {
	amountFloat := new(big.Float).SetInt(amount)
	divisor := big.NewFloat(1e6)
	result, _ := new(big.Float).Quo(amountFloat, divisor).Float64()
	return result
}
