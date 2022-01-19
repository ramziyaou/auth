package response

import (
	"auth/domain"
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func RespondWithError(ctx *fasthttp.RequestCtx, status int, message string) {
	ctx.SetStatusCode(status)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      false,
			Message: message,
		},
	)
}

func RespondInternalServerError(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      false,
			Message: "something went wrong, please try again later",
		},
	)
}

func ResponseJSON(ctx *fasthttp.RequestCtx, data string) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      true,
			Message: data,
		},
	)
}

func ResponseTransaction(ctx *fasthttp.RequestCtx, account string) { //ts []domain.Transaction) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      true,
			Message: account,
			//Transactions: ts,
		},
	)
}
