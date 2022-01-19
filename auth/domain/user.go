package domain

type User struct {
	ID       int    `json:"id"`
	Ts       string `json:"ts"`
	IIN      string `json:"iin"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}
