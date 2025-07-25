package constants

import "time"

const (
	APIV1              string        = "v1"
	BackFillBlocksSize int64         = 1000
	InitSchema         string        = "migrations/001_init.sql"
	TimeOut            time.Duration = time.Second * 10
	Deposit            string        = "deposit"
	Withdrawal         string        = "withdrawal"
	Credit             string        = "credit"
	Debit              string        = "debit"
)
