package delivery

import (
	"fmt"
	"net"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type postData struct {
	key   string
	value string
}

var testTable = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"get-info", "/info", "GET", []postData{}, fasthttp.StatusOK},
	{"get-login", "/login", "GET", []postData{}, fasthttp.StatusOK},
	{"get-signup", "/signup", "GET", []postData{}, fasthttp.StatusOK},
	{"get-update", "/update", "GET", []postData{}, fasthttp.StatusOK},
	{"get-topup", "/topup", "GET", []postData{}, fasthttp.StatusOK},
	{"get-login", "/transfer", "GET", []postData{}, fasthttp.StatusOK},
	{"get-logout", "/logout", "GET", []postData{}, fasthttp.StatusSeeOther},
	{"get-home", "/", "GET", []postData{}, fasthttp.StatusOK},
	{"get-getTransactions", "/transactions?account=KZT0000000001", "GET", []postData{}, fasthttp.StatusOK},
	{"post-addWallet", "/add", "POST", []postData{}, fasthttp.StatusOK},
	{"post-login", "/login", "POST", []postData{
		{key: "login", value: "user"},
		{key: "password", value: PASSWORD},
	}, fasthttp.StatusOK},
	{"post-signup", "/signup", "POST", []postData{
		{key: "iin", value: "980124450084"},
		{key: "login", value: "user"},
		{key: "password", value: "password "},
	}, fasthttp.StatusOK},
	{"post-topup", "/topup", "POST", []postData{
		{key: "accountno", value: "KZT0000000001"},
		{key: "amount", value: "123"},
	}, fasthttp.StatusOK},
	{"post-transfer", "/transfer", "POST", []postData{
		{key: "from", value: "KZT0000000001"},
		{key: "to", value: "KZT0000000002"},
		{key: "amount", value: "1"},
	}, fasthttp.StatusOK},
}

func TestUserHandlers(t *testing.T) {
	r := getRoutes()

	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		_ = ln.Close()
	}()

	s := &fasthttp.Server{
		Handler: r,
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
	//req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://test.com")
	// valid thru Mon Sep 08 2053 19:21:28 GMT+0600
	access, refresh, err := GenerateTestTokens("910815450350")
	if err != nil {
		t.Error("Couldn't generate token", err)
		return
	}
	req.Header.SetCookie("access", access)
	req.Header.SetCookie("refresh", refresh)
	for _, tt := range testTable {
		fmt.Println("Testing", tt.name, "******************************************************************************************")
		if tt.method == "GET" {
			req.Header.SetMethod(fasthttp.MethodGet)
			req.SetRequestURI("http://test.com" + tt.url)
			// req.SetBodyString("test")

			err := c.Do(req, res)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != tt.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, res.StatusCode())
			}
			// status, _, err := c.Get(nil, "http://test.com"+tt.url)
			// if err != nil {
			// 	t.Log(err)
			// 	t.Fatal(err)
			// }
			// if status != tt.expectedStatusCode {
			// 	t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, status)
			// }
		} else {
			URIWithArgs := URI + tt.url + "?"
			for i, p := range tt.params {
				URIWithArgs += fmt.Sprintf("%s=%s", p.key, p.value)
				if i != len(tt.params)-1 {
					URIWithArgs += "&"
				}
			}
			req.Header.SetMethod(fasthttp.MethodPost)
			req.SetRequestURI(URIWithArgs)
			// req.SetBodyString("test")

			err := c.Do(req, res)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != tt.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, res.StatusCode())
			}
			//args := fasthttp.AcquireArgs()
			// defer fasthttp.ReleaseArgs(args)
			// for _, p := range tt.params {
			// 	args.Add(p.key, p.value)
			// }
			// status, _, err := c.Post(nil, "http://test.com"+tt.url, args)
			// if err != nil {
			// 	t.Errorf("Error sending post request: %v", err)
			// 	return
			// }
			// if status != tt.expectedStatusCode {
			// 	t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, status)
			// }
		}
	}

	//req.Header.SetCookie("access", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6ZmFsc2UsImV4cCI6MjY0MDk1MDQ4OCwiaWF0IjoxNjQwMDI0ODk5LCJpaW4iOiI5MTA4MTU0NTAzNTAiLCJ1c2VybmFtZSI6ImFzcyIsInVzZXJ0cyI6IjIwMjEtMTItMzEgMTk6MzY6MzYifQ.e7Ru_Au4mbKklptDKQ20hFY7K025481NSL4-iUOza5w")
	// req.SetBodyString("test")

	if res.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("unexpected status code %d. Expecting %d", res.StatusCode(), fasthttp.StatusOK)
	}
}

var testTableErr = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
	IIN                string
	hasToken           bool
	wrongToken         bool
}{
	{"get-update-no token", "/update", "GET", []postData{}, fasthttp.StatusSeeOther, "", false, false},
	{"get-update-wrong token", "/update", "GET", []postData{}, fasthttp.StatusSeeOther, "980124450084", true, true},
	{"get-update-wrong token", "/update", "GET", []postData{}, fasthttp.StatusSeeOther, "980124450084", true, false},
	{"get-update-nonexistent user", "/update", "GET", []postData{}, fasthttp.StatusSeeOther, "nonexistent", true, false},
	{"get-update-sth wrong", "/update", "GET", []postData{}, fasthttp.StatusInternalServerError, "sthwrong", true, false},
	{"get-info some err", "/info", "GET", []postData{}, fasthttp.StatusInternalServerError, "wrong", true, false},
	{"get-transactions-no acc", "/transactions", "GET", []postData{}, fasthttp.StatusBadRequest, "", true, false},
	{"get-transactions-some err", "/transactions", "GET", []postData{
		{key: "account", value: "err"},
	}, fasthttp.StatusInternalServerError, "", true, false},
	{"post-signup-wrong iin", "/signup", "POST", []postData{
		{key: "iin", value: "980124450044"},
		{key: "login", value: "user"},
		{key: "password", value: "password "},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-wrong iin", "/signup", "POST", []postData{
		{key: "iin", value: "980124050084"},
		{key: "login", value: "user"},
		{key: "password", value: "password "},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-wrong iin", "/signup", "POST", []postData{
		{key: "iin", value: "-98012445004"},
		{key: "login", value: "user"},
		{key: "password", value: "password "},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-wrong iin", "/signup", "POST", []postData{
		{key: "iin", value: "9801244500444"},
		{key: "login", value: "user"},
		{key: "password", value: "password "},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-wrong username", "/signup", "POST", []postData{
		{key: "iin", value: "980124450084"},
		{key: "login", value: "логин"},
		{key: "password", value: "password "},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-wrong password", "/signup", "POST", []postData{
		{key: "iin", value: "980124450084"},
		{key: "login", value: "user"},
		{key: "password", value: "лыодвф"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-wrong password no special char", "/signup", "POST", []postData{
		{key: "iin", value: "980124450084"},
		{key: "login", value: "user"},
		{key: "password", value: "passsword"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-duplicate user", "/signup", "POST", []postData{
		{key: "iin", value: "980124450084"},
		{key: "login", value: "exists"},
		{key: "password", value: "password "},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-signup-duplicate user", "/signup", "POST", []postData{
		{key: "iin", value: "980124450084"},
		{key: "login", value: "other"},
		{key: "password", value: "password "},
	}, fasthttp.StatusInternalServerError, "", true, false},
	{"post-topup wrong amt", "/topup", "POST", []postData{
		{key: "accountno", value: "KZT0000000001"},
		{key: "amount", value: "-11"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-topup wrong amt", "/topup", "POST", []postData{
		{key: "accountno", value: "KZT0000000001"},
		{key: "amount", value: "0"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-topup wrong amt", "/topup", "POST", []postData{
		{key: "accountno", value: "KZT0000000001"},
		{key: "amount", value: "nb"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-topup wrong acc", "/topup", "POST", []postData{
		{key: "accountno", value: "KZTO000000001"},
		{key: "amount", value: "11"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-topup some err", "/topup", "POST", []postData{
		{key: "accountno", value: "KZT0000000001"},
		{key: "amount", value: "111"},
	}, fasthttp.StatusInternalServerError, "err", true, false},
	{"post-transfer wrong from acc", "/transfer", "POST", []postData{
		{key: "from", value: ""},
		{key: "to", value: "KZT0000000001"},
		{key: "amount", value: "111"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-transfer wrong to acc", "/transfer", "POST", []postData{
		{key: "from", value: "KZT0000000001"},
		{key: "to", value: ""},
		{key: "amount", value: "111"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-transfer wrong amt", "/transfer", "POST", []postData{
		{key: "from", value: "KZT0000000001"},
		{key: "to", value: "KZT0000000002"},
		{key: "amount", value: "-111"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-transfer same from and to acc", "/transfer", "POST", []postData{
		{key: "from", value: "KZT0000000001"},
		{key: "to", value: "KZT0000000001"},
		{key: "amount", value: "111"},
	}, fasthttp.StatusBadRequest, "", true, false},
	{"post-transfer some err", "/transfer", "POST", []postData{
		{key: "from", value: "KZT0000000001"},
		{key: "to", value: "KZT0000000001"},
		{key: "amount", value: "111"},
	}, fasthttp.StatusBadRequest, "err", true, false},
}

func TestUserHandlersError(t *testing.T) {
	r := getRoutes()

	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		_ = ln.Close()
	}()

	s := &fasthttp.Server{
		Handler: r,
	}

	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	//req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://test.com")
	// valid thru Mon Sep 08 2053 19:21:28 GMT+0600

	for _, tt := range testTableErr {
		fmt.Println("Testing", tt.name, "******************************************************************************************")
		if tt.hasToken {
			var access, refresh string
			if tt.wrongToken {
				accessT, refreshT, err := GenerateWrongToken()
				if err != nil {
					t.Error("Couldn't generate token", err)
					return
				}
				access, refresh = accessT, refreshT
			} else {
				accessT, refreshT, err := GenerateTestTokens(tt.IIN)
				if err != nil {
					t.Error("Couldn't generate token", err)
					return
				}
				access, refresh = accessT, refreshT
			}

			req.Header.SetCookie("access", access)
			req.Header.SetCookie("refresh", refresh)
		}
		if tt.method == "GET" {
			req.Header.SetMethod(fasthttp.MethodGet)
			req.SetRequestURI("http://test.com" + tt.url)
			if len(tt.params) > 0 {
				URIWithArgs := URI + tt.url + "?"
				for i, p := range tt.params {
					URIWithArgs += fmt.Sprintf("%s=%s", p.key, p.value)
					if i != len(tt.params)-1 {
						URIWithArgs += "&"
					}
				}
				req.SetRequestURI(URIWithArgs)
			}

			// req.SetBodyString("test")

			err := c.Do(req, res)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != tt.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, res.StatusCode())
			}
			// status, _, err := c.Get(nil, "http://test.com"+tt.url)

		} else {
			URIWithArgs := URI + tt.url + "?"
			for i, p := range tt.params {
				URIWithArgs += fmt.Sprintf("%s=%s", p.key, p.value)
				if i != len(tt.params)-1 {
					URIWithArgs += "&"
				}
			}
			req.Header.SetMethod(fasthttp.MethodPost)
			req.SetRequestURI(URIWithArgs)
			// req.SetBodyString("test")

			err := c.Do(req, res)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != tt.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, res.StatusCode())
			}
		}
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}
}
