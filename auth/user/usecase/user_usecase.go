package usecase

import (
	"auth/domain"
	"auth/user/repository"
	"time"
)

type TopupPageUsecase interface {
	GetWallets(string) ([]string, error)
}

type topupPageUsecaseImpl struct {
	api repository.APIInterface
}

// GetWallets Gets user accounts as string slice
func (uc *topupPageUsecaseImpl) GetWallets(token string) ([]string, error) {
	walletList, err := uc.api.GetWalletList(token)
	if err != nil {
		return nil, err
	}
	return walletList, nil
}

// NewTopupPageUsecase returns new TopupPageUsecase
func NewTopupPageUsecase(api repository.APIInterface) TopupPageUsecase {
	return &topupPageUsecaseImpl{
		api: api,
	}
}

type TransferPageUsecase interface {
	GetWallets(string) ([]string, error)
}

type transferPageUsecaseImpl struct {
	api repository.APIInterface
}

// GetWallets retieves all user accounts
func (uc *transferPageUsecaseImpl) GetWallets(token string) ([]string, error) {
	walletList, err := uc.api.GetWalletList(token)
	if err != nil {
		return nil, err
	}
	return walletList, nil
}

// NewTransferPageUsecase returns new TransferPageUsecase
func NewTransferPageUsecase(api repository.APIInterface) TransferPageUsecase {
	return &transferPageUsecaseImpl{
		api: api,
	}
}

type UpdateTokenUsecase interface {
	FindToken(IIN, token string) bool
	GetUser(IIN string) (*domain.User, error)
	InsertToken(IIN, token string, refreshTtl time.Duration) error
}

type updateTokenUsecaseImpl struct {
	cacheConn repository.CacheInterface
	dbConn    repository.DBInterface
}

// InsertToken inserts token
func (uc *updateTokenUsecaseImpl) InsertToken(IIN, token string, refreshTtl time.Duration) error {
	return uc.cacheConn.InsertToken(IIN, token, refreshTtl)
}

// GetUser gets user by IIN
func (uc *updateTokenUsecaseImpl) GetUser(IIN string) (*domain.User, error) {
	return uc.dbConn.GetUserByIIN(IIN)
}

// FindToken looks for refresh token in redis
func (uc *updateTokenUsecaseImpl) FindToken(IIN, token string) bool {
	value, err := uc.cacheConn.FindToken(IIN, token)
	if err != nil {
		return false
	}
	return token == value
}

// NewUpdateTokenUsecase returns new UpdateTokenUsecase
func NewUpdateTokenUsecase(c repository.CacheInterface, db repository.DBInterface) UpdateTokenUsecase {
	return &updateTokenUsecaseImpl{
		cacheConn: c,
		dbConn:    db,
	}
}

type LoginUsecase interface {
	GetUser(string) (*domain.User, error)
	InsertToken(IIN, token string, refreshTtl time.Duration) error
}

type loginUsecaseImpl struct {
	cacheConn repository.CacheInterface
	dbConn    repository.DBInterface
}

// GetUser gets user by username
func (uc *loginUsecaseImpl) GetUser(username string) (*domain.User, error) {
	user, err := uc.dbConn.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// InsertToken inserts new refresh token in redis
func (uc *loginUsecaseImpl) InsertToken(IIN, token string, refreshTtl time.Duration) error {
	return uc.cacheConn.InsertToken(IIN, token, refreshTtl)
}

// NewLoginUsecase return new LoginUsecase
func NewLoginUsecase(c repository.CacheInterface, db repository.DBInterface) LoginUsecase {
	return &loginUsecaseImpl{
		cacheConn: c,
		dbConn:    db,
	}
}

type AddWalletUsecase interface {
	AddWallet(string) (string, error)
}

type addWalletUsecaseImpl struct {
	api repository.APIInterface
}

// AddWallet creates new account
func (uc *addWalletUsecaseImpl) AddWallet(token string) (string, error) {
	account, err := uc.api.AddWallet(token)
	if err != nil {
		return "", err
	}
	return account, nil
}

// NewAddWalletUsecase returns new AddWalletUsecase(
func NewAddWalletUsecase(api repository.APIInterface) AddWalletUsecase {
	return &addWalletUsecaseImpl{
		api: api,
	}
}

type SignupUsecase interface {
	AddUser(IIN, username, password string) error
}

type signupUsecaseImpl struct {
	dbConn repository.DBInterface
}

// AddUser adds new user to DB
func (uc *signupUsecaseImpl) AddUser(IIN, username, password string) error {
	return uc.dbConn.AddUser(IIN, username, password)
}

// NewSignupUsecase returns new SignupUsecase
func NewSignupUsecase(db repository.DBInterface) SignupUsecase {
	return &signupUsecaseImpl{
		dbConn: db,
	}
}

type GetInfoUsecase interface {
	GetUserInfo(string) (*domain.User, error)
	GetWalletInfo(IIN, token string) ([]domain.Wallet, error)
}

type getInfoUsecaseImpl struct {
	api    repository.APIInterface
	dbConn repository.DBInterface
}

// GetUserInfo retrieves user data from DB
func (uc *getInfoUsecaseImpl) GetUserInfo(username string) (*domain.User, error) {
	user, err := uc.dbConn.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetWalletInfo retrieves account data from api
func (uc *getInfoUsecaseImpl) GetWalletInfo(IIN, token string) ([]domain.Wallet, error) {
	wallets, err := uc.api.GetWallets(IIN, token)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

// NewGetInfoUsecase returns new GetInfoUsecase
func NewGetInfoUsecase(api repository.APIInterface, db repository.DBInterface) GetInfoUsecase {
	return &getInfoUsecaseImpl{
		api:    api,
		dbConn: db,
	}
}

type GetTransactionsUsecase interface {
	GetTransactions(token, account string) ([]domain.Transaction, error)
}

type getTransactionsUsecaseImpl struct {
	api repository.APIInterface
}

func (uc *getTransactionsUsecaseImpl) GetTransactions(token, account string) ([]domain.Transaction, error) {
	transactions, err := uc.api.GetTransactions(token, account)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func NewGetTransactionsUsecase(api repository.APIInterface) GetTransactionsUsecase {
	return &getTransactionsUsecaseImpl{
		api: api,
	}
}

type TopupUsecase interface {
	TopUp(IIN, account, amount, token string) ([]byte, int, error)
}

type topupUsecaseImpl struct {
	api repository.APIInterface
}

func (uc *topupUsecaseImpl) TopUp(IIN, account, amount, token string) ([]byte, int, error) {
	resp, status, err := uc.api.TopUp(IIN, account, amount, token)
	if err != nil {
		return nil, 0, err
	}
	return resp, status, nil
}

func NewTopupUsecase(api repository.APIInterface) TopupUsecase {
	return &topupUsecaseImpl{
		api: api,
	}
}

type TransferUsecase interface {
	Transfer(IIN, from, to, amount, token string) ([]byte, int, error)
}

type transferUsecaseImpl struct {
	api repository.APIInterface
}

func (uc *transferUsecaseImpl) Transfer(IIN, from, to, amount, token string) ([]byte, int, error) {
	resp, status, err := uc.api.Transfer(IIN, from, to, amount, token)
	if err != nil {
		return nil, 0, err
	}
	return resp, status, nil
}

func NewTransferUsecase(api repository.APIInterface) TransferUsecase {
	return &transferUsecaseImpl{
		api: api,
	}
}
