package delivery

import (
	"auth/domain"
	"auth/myerrors"
	"auth/user/repository"
	"auth/user/usecase"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
)

const (
	ACCESS_SECRET   = "testingaccess"
	REFRESH_SECRET  = "testingrefresh"
	URI             = "http://test.com"
	HASHED_PASSWORD = "$2a$10$YVWoFp84S4F7TkIkV2KhguNmQ4bkQRhN14fz.MeocFLOO7XBkLxH."
	PASSWORD        = "password "
)

var refreshToken string

func TestMain(m *testing.M) {
	os.Setenv("ACCESS_SECRET", ACCESS_SECRET)
	os.Setenv("REFRESH_SECRET", REFRESH_SECRET)
	os.Exit(m.Run())
}

func getRoutes() fasthttp.RequestHandler {
	r := fasthttprouter.New()

	dbConn, err := NewMySQLDBInterface()
	if err != nil {
		log.Fatalf("Db interface create error: %v", err)
	}
	api := NewWalletAPIInterface()
	redis, err := NewRedisCacheInterface()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	log.Println("DB, API, cache", dbConn, api, redis)
	updateTokenusecase := usecase.NewUpdateTokenUsecase(redis, dbConn)
	loginUsecase := usecase.NewLoginUsecase(redis, dbConn)
	addWalletUsecase := usecase.NewAddWalletUsecase(api)
	getInfoUsecase := usecase.NewGetInfoUsecase(api, dbConn)
	signupUsecase := usecase.NewSignupUsecase(dbConn)
	topupUsecase := usecase.NewTopupUsecase(api)
	transferUsecase := usecase.NewTransferUsecase(api)
	topupPageUsecase := usecase.NewTopupPageUsecase(api)
	transferPageUsecase := usecase.NewTransferPageUsecase(api)
	getTransactionsUsecase := usecase.NewGetTransactionsUsecase(api)
	tc, err := CreateTestTemplateCache()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	NewHomePageHandler(r, tc["home.page.html"])
	NewLogoutHandler(r)
	NewGetUserInfoHandler(r, getInfoUsecase, tc["info.page.html"])
	NewGetTransactionsHandler(r, getTransactionsUsecase, tc["transactions.page.html"])
	NewLoginPageHandler(r, tc["login.page.html"])
	NewLoginHandler(r, loginUsecase)
	NewSignupPageHandler(r, tc["signup.page.html"])
	NewSignupHandler(r, signupUsecase)
	NewTopupPageHandler(r, tc["topup.page.html"], topupPageUsecase)
	NewTopupHandler(r, topupUsecase)
	NewTransferPageHandler(r, tc["transfer.page.html"], transferPageUsecase)
	NewTransferHandler(r, transferUsecase)
	NewUpdateHandler(r, updateTokenusecase, tc["update.page.html"])
	NewAddWalletHandler(r, addWalletUsecase)
	return r.Handler
}

var pathToTemplates = "../../cmd/templates/"

var functions = template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
}

func GenerateWrongToken() (string, string, error) {
	accessTokenExp := time.Now().Add(20 * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":     false,
		"exp":       accessTokenExp,
		"iat":       1640024899,
		"iin":       "910815450350",
		"createdAt": "2021-12-31 19:36:36",
	})

	accessTokenString, err := accessToken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return "", "", err
	}

	refreshTokenString, err := accessToken.SignedString([]byte(REFRESH_SECRET))

	if err != nil {
		return "", "", err
	}
	refreshToken = refreshTokenString
	return accessTokenString, refreshTokenString, nil
}

func GenerateTestTokens(IIN string) (string, string, error) {
	accessTokenExp := time.Now().Add(20 * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":     false,
		"exp":       accessTokenExp,
		"iat":       1640024899,
		"iin":       IIN,
		"username":  "sth",
		"createdAt": "2021-12-31 19:36:36",
	})

	accessTokenString, err := accessToken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return "", "", err
	}

	refreshTokenString, err := accessToken.SignedString([]byte(REFRESH_SECRET))

	if err != nil {
		return "", "", err
	}
	refreshToken = refreshTokenString
	return accessTokenString, refreshTokenString, nil
}

// CreateTestTemplateCache creates a template cache as a map
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	fmt.Println("create tc hit")
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s*.layout.html", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}
		fmt.Println(name)
		myCache[name] = ts
	}
	fmt.Println(myCache)
	return myCache, nil
}

type testDB struct{}

func (m *testDB) GetUser(username string) (*domain.User, error) {
	return &domain.User{Password: HASHED_PASSWORD}, nil
}

func (m *testDB) AddUser(IIN, username, password string) error {
	if username == "exists" {
		return myerrors.ErrDuplicateUser
	}
	if username == "other" {
		return fmt.Errorf("some other err")
	}
	return nil
}

func (m *testDB) GetUserByIIN(IIN string) (*domain.User, error) {
	if IIN == "nonexistent" {
		return nil, myerrors.ErrUserNotFound
	}
	if IIN == "sthwrong" {
		return nil, fmt.Errorf("Some other error")
	}
	return &domain.User{Password: HASHED_PASSWORD}, nil
}

func (m *testDB) Close() {}

func NewMySQLDBInterface() (repository.DBInterface, error) {
	return &testDB{}, nil
}

type testAPI struct{}

func (w *testAPI) GetTransactions(token, account string) ([]domain.Transaction, error) {
	if account == "err" {
		return nil, fmt.Errorf("some err")
	}
	return []domain.Transaction{}, nil
}

func (w *testAPI) GetWallets(IIN, token string) ([]domain.Wallet, error) {
	var wallets []domain.Wallet
	if IIN == "wrong" {
		return nil, fmt.Errorf("some err")
	}
	return wallets, nil
}

func (w *testAPI) AddWallet(token string) (string, error) {
	return "ss", nil
}

func (w *testAPI) GetWalletList(token string) ([]string, error) {
	log.Println("api hit")
	return []string{}, nil
}

func (w *testAPI) TopUp(IIN, account, amount, token string) ([]byte, int, error) {
	resp := domain.Response{OK: true}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, 0, err
	}
	if IIN == "err" {
		return respBytes, 0, fmt.Errorf("some err")
	}
	return respBytes, fasthttp.StatusOK, nil
}

func (w *testAPI) Transfer(IIN, from, to, amount, token string) ([]byte, int, error) {
	resp := domain.Response{
		OK: true,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, 0, err
	}
	if IIN == "err" {
		return respBytes, 0, fmt.Errorf("some err")
	}
	return respBytes, fasthttp.StatusOK, nil
}

func NewWalletAPIInterface() repository.APIInterface {
	return &testAPI{}
}

type testCache struct{}

func (r *testCache) InsertToken(IIN, token string, refreshTtl time.Duration) error {
	if IIN == "inserterr" {
		return fmt.Errorf("err")
	}
	return nil
}

func (r *testCache) FindToken(IIN, token string) (string, error) {
	if IIN == "980124450084" {
		return "", fmt.Errorf("Token not foumnd")
	}
	return refreshToken, nil
}

func NewRedisCacheInterface() (repository.CacheInterface, error) {
	return &testCache{}, nil
}
