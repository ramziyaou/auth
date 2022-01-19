package domain

type Response struct {
	OK           bool          `json:"ok"`
	Message      string        `json:"message"`
	WalletList   []string      `json:"walletList"`
	Wallets      []Wallet      `json:"wallets"`
	Transactions []Transaction `json:"transactions"`
}
