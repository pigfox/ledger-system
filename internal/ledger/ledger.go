package ledger

func IsBalanced(entries []LedgerEntry) bool {
	var debitSum, creditSum float64
	for _, e := range entries {
		if e.Direction == "debit" {
			debitSum += e.Amount
		} else {
			creditSum += e.Amount
		}
	}
	return debitSum == creditSum
}
