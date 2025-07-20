package db

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserAddress struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

type TransactionRequest struct {
	UserID   int     `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	TxHash   string  `json:"tx_hash,omitempty"`
	Type     string  `json:"type"`
}

type TransferRequest struct {
	FromUserID int     `json:"from_user_id"`
	ToUserID   int     `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
