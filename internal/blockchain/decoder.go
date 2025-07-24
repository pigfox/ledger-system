package blockchain

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
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
