package delivery

import (
	"auth/domain"
	"auth/myerrors"
	"auth/user/delivery/middleware"
	"auth/user/delivery/render"
	"auth/user/delivery/response"
	"auth/user/usecase"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

const (
	InternalServerErrorMessage = "что-то пошло не так, попробуйте позже"
	NoAccount                  = "предоставьте номер счета"
)

// GenerateTokens generates access and refresh tokens and sets them as cookies and ctx.UserValue
func GenerateTokens(ctx *fasthttp.RequestCtx, user *domain.User, makeRefresh bool) (string, string, error) {
	log.Println("INFO|Starting to generate tokens")
	var refreshTokenString string
	accessSecret, refreshSecret, err := middleware.GetSecretFromCtx(ctx)
	if err != nil {
		return "", "", err
	}
	ctx.SetUserValue("accessSecret", accessSecret)
	ctx.SetUserValue("refreshSecret", refreshSecret)

	accessTokenExp := time.Now().Add(20 * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iin":       user.IIN,
		"username":  user.Username,
		"createdAt": user.Ts,
		"admin":     user.IsAdmin,
		"exp":       accessTokenExp,
	})

	accessTokenString, err := accessToken.SignedString([]byte(accessSecret))

	if err != nil {
		return "", "", err
	}
	if makeRefresh {
		refreshTokenExp := time.Now().Add(10 * time.Minute).Unix()

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iin":       user.IIN,
			"username":  user.Username,
			"createdAt": user.Ts,
			"admin":     false,
			"exp":       refreshTokenExp,
		})

		refreshTokenString, err = refreshToken.SignedString([]byte(refreshSecret))
		if err != nil {
			return "", "", err
		}

		refreshCookie := makeCookie("refresh", refreshTokenString, 3600)
		ctx.Response.Header.SetCookie(refreshCookie)
	}
	accessCookie := makeCookie("access", accessTokenString, 3600)
	ctx.Response.Header.SetCookie(accessCookie)
	IINCookie := makeCookie("iin", user.IIN, 25)
	ctx.Response.Header.SetCookie(IINCookie)
	ctx.SetUserValue("access", accessTokenString)

	log.Printf("INFO|Generated cookies\nAccess:%s\nRefresh:%s", accessTokenString, refreshTokenString)
	return accessTokenString, refreshTokenString, nil
}

// makeCookie makes *fasthttp.Cookkies
func makeCookie(key, value string, age int) *fasthttp.Cookie {
	authCookie := fasthttp.Cookie{}
	authCookie.SetKey(key)
	authCookie.SetValue(value)
	authCookie.SetMaxAge(age)
	authCookie.SetHTTPOnly(true)
	authCookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	return &authCookie
}

type UpdateHandler struct {
	ucUpdate usecase.UpdateTokenUsecase
	t        *template.Template
}

// UpdateToken handles update of access token
func (h *UpdateHandler) UpdateToken(ctx *fasthttp.RequestCtx) {
	fmt.Println("INFO|Update endpoint hit, getting refreshtoken")
	// isAdmin, ok := ctx.Value("admin").(bool)
	// if !ok {
	// 	log.Println("ERROR|Failed to retrieve role from ctx")
	// 	response.RespondInternalServerError(ctx)
	// 	return
	// }
	refreshToken, err := middleware.ExtractToken(ctx, false)
	var message string

	if err != nil {
		log.Println("ERROR|Updating token:", err)
		ctx.SetStatusCode(fasthttp.StatusSeeOther)
		ctx.Response.Header.Add("Location", "/login")
		return
	}
	log.Println("INFO|Updatetoken: extracted refreshtoken", refreshToken)
	IIN, _, err := middleware.ParseToken(ctx, refreshToken, false)
	if err != nil {
		log.Printf("ERROR|Parse refresh token error: %v", err)
		ctx.SetStatusCode(fasthttp.StatusSeeOther)
		ctx.Response.Header.Add("Location", "/login")
		return
	}
	log.Println("INFO|Updatetoken: parsed refreshtoken")

	// validate token through Redis
	ok := h.ucUpdate.FindToken(IIN, refreshToken)
	if !ok {
		log.Printf("ERROR|Getting refresh token failed")
		ctx.SetStatusCode(fasthttp.StatusSeeOther)
		ctx.Response.Header.Add("Location", "/login")
		return
	}
	log.Println("INFO|Updatetoken:Found and validated refresh token on Redis")
	user, err := h.ucUpdate.GetUser(IIN)
	if err != nil {
		log.Println("ERROR|Couldn't get user to generate token:", err)
		if err == myerrors.ErrUserNotFound {
			ctx.SetStatusCode(fasthttp.StatusSeeOther)
			ctx.Response.Header.Add("Location", "/login")
			return
		}

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		message = "couldn't find token, please try login page"
		render.RenderTemplate(ctx, fasthttp.StatusInternalServerError, h.t, message)
		return
	}
	// if isAdmin {
	// 	user.IIN = "0"
	// 	user.IsAdmin = true
	// }
	fmt.Println("INFO|Generating new refresh token")
	access, refresh, err := GenerateTokens(ctx, user, false)
	if err != nil {
		log.Println("ERROR|Parse refresh token error:", err)
		ctx.SetStatusCode(fasthttp.StatusSeeOther)
		ctx.Response.Header.Add("Location", "/login")
		return
	}

	log.Println("INFO|Updated access")
	message = fmt.Sprintf("access: %s\n\nrefresh: %s\n", access, refresh)

	render.RenderTemplate(ctx, fasthttp.StatusOK, h.t, message)
}

// NewUpdateHandler sets /update route
func NewUpdateHandler(r *fasthttprouter.Router, ucUpdate usecase.UpdateTokenUsecase, t *template.Template) {
	handler := &UpdateHandler{
		ucUpdate: ucUpdate,
		t:        t,
	}
	r.GET("/update", middleware.SecretMiddleware(handler.UpdateToken))
}

type AddWalletHandler struct {
	uc usecase.AddWalletUsecase
}

// AddWallet handles creation of new wallets
func (h *AddWalletHandler) AddWallet(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|AddWallet hit")
	token, ok := ctx.Value("access").(string)
	if !ok {
		log.Println("ERROR|Failed to retrieve token for add wallet handler")
		response.RespondInternalServerError(ctx)
		return
	}
	account, err := h.uc.AddWallet(token)
	if err != nil {
		log.Println("ERROR|Couldn't add wallet", err)
		response.RespondInternalServerError(ctx)
		return
	}
	response.ResponseJSON(ctx, "Created new account under "+account)
}

// NewAddWalletHandler sets /add route
func NewAddWalletHandler(r *fasthttprouter.Router, uc usecase.AddWalletUsecase) {
	handler := &AddWalletHandler{
		uc: uc,
	}
	r.POST("/add", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.AddWallet)))
}

type LoginPageHandler struct {
	t *template.Template
}

// LoginPageHandler serves static login page
func (h *LoginPageHandler) LoginPage(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|LoginPage hit")
	ctx.Response.Header.SetContentType("text/html")
	if err := h.t.Execute(ctx, nil); err != nil {
		log.Println("ERROR|LoginPage handler:", err)
		response.RespondInternalServerError(ctx)
		return
	}
}

// NewLoginPageHandler sets /login GET route
func NewLoginPageHandler(r *fasthttprouter.Router, t *template.Template) {
	handler := &LoginPageHandler{
		t: t,
	}
	r.GET("/login", middleware.SecretMiddleware(handler.LoginPage))
}

type LoginHandler struct {
	uc usecase.LoginUsecase
}

// LogIn handles provided username and password
func (h *LoginHandler) LogIn(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|LogIn hit")
	username, password := extractCredential(ctx)
	fmt.Println("IIN,login,pass:", username, password)
	user, err := h.uc.GetUser(username)
	if err != nil {
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "invalid user, try again or sign up")
		log.Println("ERROR|Couldn't find user", err)
		return
	}
	log.Println("INFO|Succesfully retrieved user:", user)
	hashedPassword := user.Password
	ctx.SetUserValue("user", user)
	fmt.Println("here's USER IN LOGIN:", user)

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		log.Println("ERROR|Invalid password:", err)
		response.RespondWithError(ctx, fasthttp.StatusUnauthorized, "invalid password")
		return
	}
	if username == "admin" {
		user.IsAdmin = true
	}
	_, refresh, err := GenerateTokens(ctx, user, true)
	if err != nil {
		log.Println("ERROR|Login handler:", err)
		response.RespondInternalServerError(ctx)
		return
	}

	if err := h.uc.InsertToken(user.IIN, refresh, 10*time.Minute); err != nil {
		log.Println("ERROR|Couldn't insert token to redis. Error:", err)
		response.RespondInternalServerError(ctx)
		return
	}
	log.Println("INFO|Successfully inserted refresh after login, redirecting to info")
	response.ResponseJSON(ctx, "Success")
}

// NewLoginHandler sets /login POST route
func NewLoginHandler(r *fasthttprouter.Router, uc usecase.LoginUsecase) {
	handler := &LoginHandler{
		uc: uc,
	}
	r.POST("/login", middleware.SecretMiddleware(handler.LogIn))
}

func extractCredential(ctx *fasthttp.RequestCtx) (login string, pass string) {
	return string(ctx.FormValue("login")), string(ctx.FormValue("password"))
}

func extractToken(ctx *fasthttp.RequestCtx, isAccess bool) (token string, err error) {
	//header := string(ctx.Request.Header.Peek("Authorization"))
	token = string(ctx.Request.Header.Cookie("access"))
	if !isAccess {
		token = string(ctx.Request.Header.Cookie("refresh"))
	}

	fmt.Println("cookie retrieving:", token)
	if token == "" {
		err = fmt.Errorf("Token cookie not found")
		return
	}
	return
}

type SignupPageHandler struct {
	t *template.Template
}

// SignupPage serves static signup page
func (h *SignupPageHandler) SignupPage(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|SignupPage hit")
	ctx.Response.Header.SetContentType("text/html")
	if err := h.t.Execute(ctx, nil); err != nil {
		log.Println("ERROR|SignupPage:", err)
		response.RespondInternalServerError(ctx)
		return
	}
}

// NewLoginHandler sets /signup GET route
func NewSignupPageHandler(r *fasthttprouter.Router, t *template.Template) {
	handler := &SignupPageHandler{
		t: t,
	}
	r.GET("/signup", middleware.SecretMiddleware(handler.SignupPage))
}

type SignupHandler struct {
	uc usecase.SignupUsecase
}

// SignUp handles signup with IIN, username and password
func (h *SignupHandler) SignUp(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|SignUp hit")
	IIN, username, password := string(ctx.FormValue("iin")), string(ctx.FormValue("login")), string(ctx.FormValue("password"))

	if !validateIIN(IIN) {
		log.Println("ERROR|Coudln't validate IIN")
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "invalid IIN")
		return
	}
	username = strings.Trim(username, " ")
	if err := validateCreds(username, password); err != nil {
		log.Println("ERROR|Coudln't validate creds")
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "invalid username or password")
		return
	}

	user := domain.User{
		IIN:      IIN,
		Username: username,
	}

	log.Println("INFO|Generating hash from password given as:", password)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Println("ERROR|Signup handler:", err)
		response.RespondInternalServerError(ctx)
		return
	}

	user.Password = string(hash)

	if err = h.uc.AddUser(user.IIN, user.Username, user.Password); err != nil {
		log.Println("ERROR|Signup handler:", err)
		if err == myerrors.ErrDuplicateUser {
			response.RespondWithError(ctx, fasthttp.StatusBadRequest, "username / IIN already exist(s)")
			return
		}
		response.RespondInternalServerError(ctx)
		return
	}
	response.ResponseJSON(ctx, "Success, click ok to redirect to login page")
}

// validateCreds validates login and password
func validateCreds(login, password string) error {
	if !strIsPrint(login) {
		return myerrors.ErrInvalidUsername
	}
	if !strIsPrint(password) || !containsSpecialChar(password) {
		return myerrors.ErrInvalidPassword
	}
	return nil
}

// validateIIN validates IIN
func validateIIN(s string) bool {
	if len(s) != 12 {
		log.Println("ERROR|Invalid IIN length")
		return false
	}
	if _, err := strconv.Atoi(s); err != nil || s[0] == '-' || s[0] == '+' {
		log.Println("ERROR|Non-numeric characters in IIN")
		return false
	}
	if s[6] == '0' || s[6] > '6' {
		log.Println("ERROR|Invalid 7th char in IIN")
		return false
	}
	var mod int
	if mod = (int(s[0]-'0') + 2*int(s[1]-'0') + 3*int(s[2]-'0') + 4*int(s[3]-'0') + 5*int(s[4]-'0') + 6*int(s[5]-'0') + 7*int(s[6]-'0') + 8*int(s[7]-'0') + 9*int(s[8]-'0') + 10*int(s[9]-'0') + 11*int(s[10]-'0')) % 11; mod == 10 {
		mod = (3*int(s[0]-'0') + 4*int(s[1]-'0') + 5*int(s[2]-'0') + 6*int(s[3]-'0') + 7*int(s[4]-'0') + 8*int(s[5]-'0') + 9*int(s[6]-'0') + 10*int(s[7]-'0') + 11*int(s[8]-'0') + int(s[9]-'0') + 2*int(s[10]-'0')) % 11
	}
	if mod != int(s[11]-'0') {
		log.Println("ERROR|Invalid last character in IIN")
		return false
	}
	return true
}

// strIsPrint reports whether the string passed consists of printable Latin character only
func strIsPrint(s string) bool {
	for _, char := range s {
		if char < 32 || char > 126 {
			return false
		}
	}
	return true
}

// containsSpecialChar reports whether the string contains a special character
func containsSpecialChar(s string) bool {
	for _, char := range s {
		if char < 48 || (char > 57 && char < 65) || (char > 90 && char < 97) || (char > 122 && char < 127) {
			return true
		}
	}
	return false
}

func NewSignupHandler(r *fasthttprouter.Router, uc usecase.SignupUsecase) {
	handler := &SignupHandler{
		uc: uc,
	}
	r.POST("/signup", middleware.SecretMiddleware(handler.SignUp))
}

type GetUserInfoHandler struct {
	uc usecase.GetInfoUsecase
	t  *template.Template
}

// GetUser retrieves account and wallet information on requested user
func (h *GetUserInfoHandler) GetUser(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|Get user hit")
	var info domain.Info
	token, ok := ctx.Value("access").(string)
	if !ok || token == "" {
		log.Println("ERROR|Couldn't get token from ctx")
		ctx.SetStatusCode(fasthttp.StatusSeeOther)
		ctx.Response.Header.Add("Location", "/login")
		return
	}
	u, ok := ctx.Value("user").(domain.User)
	if !ok {
		log.Println("ERROR|User is nil")
		render.RenderTemplate(ctx, fasthttp.StatusInternalServerError, h.t, info)
		return
	}
	log.Printf("INFO|User with IIN %s sent request", u.IIN)
	user, err := h.uc.GetUserInfo(u.Username)
	if err != nil {
		log.Println("ERROR|Error getting user info from DB:", err)
		response.RespondInternalServerError(ctx)
		render.RenderTemplate(ctx, fasthttp.StatusInternalServerError, h.t, info)
		return
	}
	wallets, err := h.uc.GetWalletInfo(u.IIN, token)
	if err != nil {
		log.Println("ERROR|Error getting wallets from DB:", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		render.RenderTemplate(ctx, fasthttp.StatusInternalServerError, h.t, info)
		return
	}
	info.User = user
	info.Wallets = wallets
	render.RenderTemplate(ctx, fasthttp.StatusOK, h.t, info)
}

// NewGetInfoHandler sets /info route
func NewGetUserInfoHandler(r *fasthttprouter.Router, uc usecase.GetInfoUsecase, t *template.Template) {
	handler := &GetUserInfoHandler{
		uc: uc,
		t:  t,
	}
	r.GET("/info", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.GetUser)))
}

type GetTransactionsHandler struct {
	uc usecase.GetTransactionsUsecase
	t  *template.Template
}

func (h *GetTransactionsHandler) GetTransactions(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|Get transactions hit")
	token, ok := ctx.Value("access").(string)
	if !ok || token == "" {
		log.Println("ERROR|Couldn't get token from ctx")
		ctx.SetStatusCode(fasthttp.StatusSeeOther)
		ctx.Response.Header.Add("Location", "/login")
		return
	}
	account := string(ctx.FormValue("account"))
	if account == "" {
		log.Println("ERROR|GetTransactions handler: Couldn't get accountNo")
		//response.RespondInternalServerError(ctx)
		if err := render.RenderTemplate(ctx, fasthttp.StatusBadRequest, h.t, domain.Info{Error: NoAccount}); err != nil {
			log.Println("ERROR|Executing template", err)
		}
		return
	}

	transactions, err := h.uc.GetTransactions(token, account)
	if err != nil {
		log.Println("ERROR|Error getting wallets", err)
		//response.RespondInternalServerError(ctx)
		if err := render.RenderTemplate(ctx, fasthttp.StatusInternalServerError, h.t, domain.Info{Error: InternalServerErrorMessage}); err != nil {
			log.Println("ERROR|Executing template", err)
		}
		return
	}
	if err := render.RenderTemplate(ctx, fasthttp.StatusOK, h.t, domain.Info{AccountNo: account, Transactions: transactions}); err != nil {
		log.Println("ERROR|Executing template", err)
	}
	log.Println("INFO|Success")
}

func NewGetTransactionsHandler(r *fasthttprouter.Router, uc usecase.GetTransactionsUsecase, t *template.Template) {
	handler := &GetTransactionsHandler{
		uc: uc,
		t:  t,
	}
	r.GET("/transactions", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.GetTransactions)))
}

func validAmt(s string) bool {
	if num, err := strconv.Atoi(s); err != nil || num <= 0 {
		fmt.Println("Invalid input")
		return false
	}
	// for _, char := range s {
	// 	if char < '0' || char > '9' {
	// 		log.Println("ERROR|Invalid input")
	// 		return false
	// 	}
	// }
	return true
}

func validAcc(s string) bool {
	if len(s) != 13 || s[0] != 'K' || s[1] != 'Z' || s[2] != 'T' {
		return false
	}
	if num, err := strconv.Atoi(s[3:]); err != nil || num < 0 {
		return false
	}
	return true
}

type TopupPageHandler struct {
	t  *template.Template
	uc usecase.TopupPageUsecase
}

func (h *TopupPageHandler) TopupPage(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|TopupPage hit")
	ctx.Response.Header.SetContentType("text/html")
	token := string(ctx.Request.Header.Cookie("access"))
	if token == "" {
		log.Println("ERROR|Couldn't get token from header")
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "couldn't find token")
		// redirect to login page?
		return
	}
	walletList, err := h.uc.GetWallets(token)
	if err != nil {
		log.Println("ERROR|TopupPage handler:", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		walletList = append(walletList, "shutdown")
		//return
	}
	if err := h.t.Execute(ctx, walletList); err != nil {
		log.Println("ERROR|TopupPage handler:", err)
		response.RespondInternalServerError(ctx)
		return
	}
}

func NewTopupPageHandler(r *fasthttprouter.Router, t *template.Template, uc usecase.TopupPageUsecase) {
	handler := &TopupPageHandler{
		t:  t,
		uc: uc,
	}
	r.GET("/topup", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.TopupPage)))
}

type TopupHandler struct {
	uc usecase.TopupUsecase
}

func extractTopupValues(ctx *fasthttp.RequestCtx) (account string, amount string, err error) {
	account = string(ctx.FormValue("accountno"))
	amount = string(ctx.FormValue("amount"))
	log.Println("INFO|Received folowing account and amount:", account, amount)
	if !validAcc(account) || !validAmt(amount) {
		err = fmt.Errorf("invalid acc or amt")
		return
	}
	return
}

func (h *TopupHandler) TopUp(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|Topup handler hit")
	account, amount, err := extractTopupValues(ctx)
	if err != nil {
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "Invalid amount")
		return
	}
	token, ok := ctx.Value("access").(string)
	if !ok {
		log.Println("ERROR|Couldn't find token")
		response.RespondInternalServerError(ctx)
		return
	}
	user, ok := ctx.Value("user").(domain.User)
	if !ok {
		log.Println("ERROR|User is nil")
		response.RespondInternalServerError(ctx)
		return
	}
	log.Println("INFO|Sending topup request from authService, token, IIN:", token, user.IIN)
	log.Printf("account:%s, amount:%s\n", account, amount)

	respBytes, status, err := h.uc.TopUp(user.IIN, account, amount, token)
	if err != nil {
		log.Println("ERROR|Coudn't get response from walletService")
		response.RespondInternalServerError(ctx)
		return
	}
	var resp domain.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		log.Println("ERROR|Getting response from walletService: ", err)
		response.RespondInternalServerError(ctx)
		return
	}
	if resp.OK {
		response.ResponseJSON(ctx, "Topped up successfully, current balance is ₸"+string(resp.Message))
		return
	}
	if status == fasthttp.StatusBadRequest {
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, resp.Message)
		return
	}
	response.RespondInternalServerError(ctx)
}

func NewTopupHandler(r *fasthttprouter.Router, uc usecase.TopupUsecase) {
	handler := &TopupHandler{uc: uc}
	r.POST("/topup", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.TopUp)))
}

type TransferPageHandler struct {
	t  *template.Template
	uc usecase.TransferPageUsecase
}

func (h *TransferPageHandler) TransferPage(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|TransferPage hit")
	ctx.Response.Header.SetContentType("text/html")
	token := string(ctx.Request.Header.Cookie("access"))
	if token == "" {
		log.Println("ERROR|Couldn't get token from ctx")
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "Couldn't find token")
		return
	}
	walletList, err := h.uc.GetWallets(token)
	if err != nil {
		log.Println("ERROR|TransferPage handler:", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		walletList = append(walletList, "shutdown")
	}
	if err := h.t.Execute(ctx, walletList); err != nil {
		log.Println("ERROR|TransferPage handler:", err)
		response.RespondInternalServerError(ctx)
		return
	}
}

func NewTransferPageHandler(r *fasthttprouter.Router, t *template.Template, uc usecase.TransferPageUsecase) {
	handler := &TransferPageHandler{
		t:  t,
		uc: uc,
	}
	r.GET("/transfer", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.TransferPage)))
}

type TransferHandler struct {
	uc usecase.TransferUsecase
}

func extractTransfervalue(ctx *fasthttp.RequestCtx) (from string, to string, amount string, err error) {

	from = string(ctx.FormValue("from"))
	to = string(ctx.FormValue("to"))
	if to == "-" {
		to = string(ctx.FormValue("other"))
	}
	log.Println("INFO|Transfering into account", to)
	amount = string(ctx.FormValue("amount"))
	if !validAmt(amount) {
		err = myerrors.ErrInvalidAmt
		return
	}
	if !validAcc(from) || !validAcc(to) {
		err = myerrors.ErrInvalidAcc
		return
	}
	if from == to {
		err = myerrors.ErrSameAccount
		return
	}
	return
}

// Transfer handles transactions between user wallets
func (h *TransferHandler) Transfer(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|Transfer endpoint hit")

	from, to, amount, err := extractTransfervalue(ctx)

	if err != nil {
		log.Println("ERROR|Extracting transfer values:", err)
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	token, ok := ctx.Value("access").(string)
	if !ok {
		fmt.Println("Couldn't find token")
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "Couldn't find token")
		return
	}
	log.Println("INFO|Transfer handler getting token", token)
	user, ok := ctx.Value("user").(domain.User)
	if !ok {
		log.Println("ERROR|User is nil")
		response.RespondInternalServerError(ctx)
		return
	}
	log.Println("INFO|Sending transfer request from authService")

	respBytes, status, err := h.uc.Transfer(user.IIN, from, to, amount, token)
	log.Println("INFO|Got the following response & status code:", string(respBytes), "\n", status)
	if err != nil {
		log.Println("ERROR|Coudn't get response from walletService")
		response.RespondInternalServerError(ctx)
		return
	}
	log.Println("Received response from walletService")
	var resp domain.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		log.Println("ERROR|Getting response from walletService: ", err)
		response.RespondInternalServerError(ctx)
		return
	}
	if resp.OK {
		response.ResponseJSON(ctx, resp.Message)
		log.Println("Everything OK")
		return
	}
	if status == fasthttp.StatusBadRequest {
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, resp.Message)
		return
	}
	response.RespondInternalServerError(ctx)
}

func NewTransferHandler(r *fasthttprouter.Router, uc usecase.TransferUsecase) {
	handler := &TransferHandler{
		uc: uc,
	}
	r.POST("/transfer", middleware.SecretMiddleware(middleware.CheckAuthMiddleware(handler.Transfer)))
}

type LogoutHandler struct{}

// LogOut handles logout by deleting token cookies
func (h *LogoutHandler) LogOut(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|LogOut hit")
	access := makeCookie("access", "", 0)
	access.SetExpire(time.Unix(0, 0))

	refresh := makeCookie("refresh", "", 0)
	refresh.SetExpire(time.Unix(0, 0))

	ctx.Response.Header.SetCookie(access)
	ctx.Response.Header.SetCookie(refresh)
	ctx.SetStatusCode(fasthttp.StatusSeeOther)
	ctx.Response.Header.Add("Location", "/")
}

// NewLogoutHandler sets /logout route
func NewLogoutHandler(r *fasthttprouter.Router) {
	handler := &LogoutHandler{}
	r.GET("/logout", handler.LogOut)
}

type HomePageHandler struct {
	t *template.Template
}

// Home handles / path
func (h *HomePageHandler) Home(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|HomePage hit")
	ctx.Response.Header.SetContentType("text/html")
	if err := h.t.Execute(ctx, nil); err != nil {
		log.Println("ERROR|LoginPage handler:", err)
		response.RespondInternalServerError(ctx)
		return
	}
}

// NewHomePageHandler sets / route
func NewHomePageHandler(r *fasthttprouter.Router, t *template.Template) {
	handler := &HomePageHandler{
		t: t,
	}
	r.GET("/", handler.Home)
}
