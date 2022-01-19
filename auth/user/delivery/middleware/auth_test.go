package middleware

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var td = map[string]string{
	"Error when parsing token": "",
	"Couldn't find IIN":        "jgq1&2w_347192",
	"Couldn't find role":       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDAwMjQ5MTksImlhdCI6MTY0MDAyNDg5OSwiaWluIjoiOTEwODE1NDUwMzUwIiwidXNlcm5hbWUiOiJhc3MiLCJ1c2VydHMiOiIyMDIxLTEyLTE5IDEwOjM2OjM2In0.ouFMo3rLUHELjupcV8yqbiu0-_jWELiZ1pE-r5kht5M",
}

const (
	ACCESS_SECRET  = "testingaccess"
	REFRESH_SECRET = "testingrefresh"
)

type Body struct {
	AccessSecret  string
	RefreshSecret string
	AccessTtl     time.Duration
	RefreshTtl    time.Duration
}

func TestSecretMiddleware(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()
	s := &fasthttp.Server{
		Handler: SecretMiddleware(func(ctx *fasthttp.RequestCtx) {
			//body := ctx.Request.Body()
			body := Body{
				AccessSecret:  ctx.UserValue("accessSecret").(string),
				RefreshSecret: ctx.UserValue("refreshSecret").(string),
				AccessTtl:     ctx.UserValue("accessTtl").(time.Duration),
				RefreshTtl:    ctx.UserValue("refreshTtl").(time.Duration),
			}

			resp, err := json.Marshal(&body)
			if err != nil {
				ctx.Write([]byte("Error"))
			}
			ctx.Write(resp) //nolint:errcheck
		}),
	}
	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://test.com")

	err := c.Do(req, res)
	if err != nil {
		t.Fatal(err)
	}
	var respBody Body
	if err := json.Unmarshal(res.Body(), &respBody); err != nil {
		t.Fatal("Unmarshal error:", err)
	}
	accessSecret, refreshSecret, accessTtl, refreshTtl := respBody.AccessSecret, respBody.RefreshSecret, respBody.AccessTtl, respBody.RefreshTtl
	if accessSecret != ACCESS_SECRET {
		t.Errorf("Unexpected accessSecret: expecting %q, got %q", ACCESS_SECRET, accessSecret)
	}
	if refreshSecret != REFRESH_SECRET {
		t.Errorf("Unexpected accessSecret: expecting %q, got %q", REFRESH_SECRET, refreshSecret)
	}
	if accessTtl != time.Duration(time.Second*20) {
		t.Errorf("Unexpected accessSecret: expecting %q, got %v", "20s", accessTtl)
	}
	if refreshTtl != time.Duration(time.Minute*10) {
		t.Errorf("Unexpected accessSecret: expecting %q, got %q", "10m0s", refreshTtl)
	}

}

func TestCheckAuthMiddleware(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()
	s := &fasthttp.Server{
		Handler: SecretMiddleware(CheckAuthMiddleware(func(ctx *fasthttp.RequestCtx) {})),
	}
	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://test.com")
	// valid through Mon Sep 08 2053 19:21:28 GMT+0600
	token, err := GenerateToken()
	if err != nil {
		t.Error("Failed to generate token", err)
		return
	}
	req.Header.SetCookie("access", token)
	err = c.Do(req, res)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode() != fasthttp.StatusOK {
		t.Errorf("unexpected status code %d. Expecting %d", res.StatusCode(), fasthttp.StatusOK)
	}
}
