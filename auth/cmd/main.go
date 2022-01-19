package main

import (
	"auth/user/delivery"
	"auth/user/delivery/render"
	"auth/user/repository/mysql"
	"auth/user/repository/redis"
	"auth/user/repository/walletservice"
	"auth/user/usecase"
	"fmt"
	"log"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/subosito/gotenv"
	"github.com/valyala/fasthttp"
)

func init() {
	gotenv.Load()
}

func main() {
	time.Local = time.FixedZone("CST", 6*3600)
	fmt.Println(time.Now())
	r := fasthttprouter.New()
	dbConn, err := mysql.NewMySQLDBInterface()
	if err != nil {
		log.Fatalf("Db interface create error: %v", err)
	}
	defer dbConn.Close()
	api := walletservice.NewWalletAPIInterface()
	redis, err := redis.NewRedisCacheInterface()
	if err != nil {
		fmt.Println(err)
		return
	}
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
	tc, err := render.CreateTemplateCache()
	if err != nil {
		fmt.Println(err)
		return
	}
	delivery.NewHomePageHandler(r, tc["home.page.html"])
	delivery.NewLogoutHandler(r)
	delivery.NewGetUserInfoHandler(r, getInfoUsecase, tc["info.page.html"])
	delivery.NewGetTransactionsHandler(r, getTransactionsUsecase, tc["transactions.page.html"])
	delivery.NewLoginPageHandler(r, tc["login.page.html"])
	delivery.NewLoginHandler(r, loginUsecase)
	delivery.NewSignupPageHandler(r, tc["signup.page.html"])
	delivery.NewSignupHandler(r, signupUsecase)
	delivery.NewTopupPageHandler(r, tc["topup.page.html"], topupPageUsecase)
	delivery.NewTopupHandler(r, topupUsecase)
	delivery.NewTransferPageHandler(r, tc["transfer.page.html"], transferPageUsecase)
	delivery.NewTransferHandler(r, transferUsecase)
	delivery.NewUpdateHandler(r, updateTokenusecase, tc["update.page.html"])
	delivery.NewAddWalletHandler(r, addWalletUsecase)
	fasthttp.ListenAndServe(":8080", r.Handler)
}
