package domain

type Info struct {
	User         *User         `json:"user"`
	AccountNo    string        `json:"transaction"`
	Transactions []Transaction `json:"transactions"`
	Wallets      []Wallet      `json:"wallets"`
	Error        string        `json:"error"`
}
