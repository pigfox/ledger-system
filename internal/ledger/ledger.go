package ledger

type Entry struct {
	Account   string
	Amount    float64
	Currency  string
	Direction string // "debit" or "credit"
}

func IsBalanced(entries []Entry) bool {
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
