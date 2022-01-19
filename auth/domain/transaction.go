package domain

type Transaction struct {
	ID     int    `json:"id"`
	Ts     string `json:"ts"`
	Type   string `json:"transfer_type"`
	From   string `json:"from_acc"`
	To     string `json:"to_acc"`
	Amount int    `json:"amount"`
}
