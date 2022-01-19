package repository

import (
	"auth/domain"
	"time"
)

type CacheInterface interface {
	InsertToken(string, string, time.Duration) error
	FindToken(string, string) (string, error)
}

type DBInterface interface {
	GetUser(username string) (*domain.User, error)
	GetUserByIIN(IIN string) (*domain.User, error)
	AddUser(IIN, username, password string) error
	Close()
}

type APIInterface interface {
	GetWallets(IIN, token string) ([]domain.Wallet, error)
	GetTransactions(token, account string) ([]domain.Transaction, error)
	GetWalletList(string) ([]string, error)
	TopUp(IIN, account, amount, token string) ([]byte, int, error)
	Transfer(IIN, from, to, amount, token string) ([]byte, int, error)
	AddWallet(token string) (string, error)
}
