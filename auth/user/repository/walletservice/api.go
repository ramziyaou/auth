package walletservice

import (
	"auth/domain"
	"auth/user/repository"
	"encoding/json"
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

type WalletAPIInterface struct {
	host   string
	client *fasthttp.HostClient
}

func (w *WalletAPIInterface) GetTransactions(token, account string) ([]domain.Transaction, error) {
	respBytes, _, err := w.doRequest("/transactions", map[string]string{"token": token, "account": account})
	if err != nil {
		return nil, err
	}

	var resp domain.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}
	log.Println(resp)
	log.Println("\n", resp.Transactions)
	return resp.Transactions, nil
}

func (w *WalletAPIInterface) GetWallets(IIN, token string) ([]domain.Wallet, error) {
	respBytes, _, err := w.doRequest("/info", map[string]string{"token": token})
	if err != nil {
		return nil, err
	}

	var resp domain.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}
	return resp.Wallets, nil
}

func (w *WalletAPIInterface) AddWallet(token string) (string, error) {
	respBytes, _, err := w.doRequest("/add", map[string]string{"token": token})
	if err != nil {
		return "", err
	}
	var resp domain.Response
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return "", err
	}
	return resp.Message, nil
}

func (w *WalletAPIInterface) GetWalletList(token string) ([]string, error) {
	respBytes, _, err := w.doRequest("/wallets", map[string]string{"token": token})
	if err != nil {
		return nil, err
	}
	var resp domain.Response
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}
	return resp.WalletList, nil
}

func (w *WalletAPIInterface) TopUp(IIN, account, amount, token string) ([]byte, int, error) {
	resp, status, err := w.doRequest("/topup", map[string]string{"iin": IIN, "account": account, "amount": amount, "token": token})
	if err != nil {
		return nil, 0, err
	}
	return resp, status, nil
}

func (w *WalletAPIInterface) doRequest(endpoint string, m map[string]string) ([]byte, int, error) {
	log.Println("INFO|do request hit")
	req := fasthttp.AcquireRequest()
	for key, value := range m {
		req.Header.Add(key, value)
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(w.host + endpoint)
	if err := w.client.DoTimeout(req, resp, time.Second*5); err != nil {
		return nil, 0, err
	}
	bodyBytes := resp.Body()
	return bodyBytes, resp.StatusCode(), nil
}

func (w *WalletAPIInterface) Transfer(IIN, from, to, amount, token string) ([]byte, int, error) {
	resp, status, err := w.doRequest("/transfer", map[string]string{"iin": IIN, "from": from, "to": to, "amount": amount, "token": token})
	if err != nil {
		return nil, 0, err
	}
	return resp, status, nil
}

func NewWalletAPIInterface() repository.APIInterface {
	return &WalletAPIInterface{
		host:   "http://host.docker.internal:8070",
		client: newClient("host.docker.internal:8070"),
	}
}

func newClient(addr string) *fasthttp.HostClient {
	return &fasthttp.HostClient{
		Addr:                     addr,
		Name:                     "WalletClient",
		NoDefaultUserAgentHeader: true,
	}
}
