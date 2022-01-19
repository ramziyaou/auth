package walletservice

import (
	"auth/domain"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTransactions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK: true,
				Transactions: []domain.Transaction{
					{ID: 1, Type: "topup", To: "KZT0000000001", Amount: 1},
					{ID: 2, Type: "transfer", From: "KZT0000000001", To: "KZT0000000002", Amount: 1},
				},
			},
		)
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	transactions, err := api.GetTransactions("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 {
		t.Errorf("Expecting 2 transactions, got %d", len(transactions))
	}
	// no transactions
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK:           true,
				Transactions: []domain.Transaction{},
			},
		)
	}))
	defer ts.Close()
	api = &WalletAPIInterface{ts.URL, newClient(ts.URL[7:])}
	transactions, err = api.GetTransactions("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expecting 0 transactions, got %d", len(transactions))
	}
}

func TestGetTransactionsErr(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte{})
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	transactions, err := api.GetTransactions("", "")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(transactions) != 0 {
		t.Errorf("Expecting 0 transactions, got %d", len(transactions))
	}
	// no response at all
	api = &WalletAPIInterface{host: "nonexistent.com", client: newClient("nonexistent.com")}
	transactions, err = api.GetTransactions("", "")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(transactions) != 0 {
		t.Errorf("Expecting 0 transactions, got %d", len(transactions))
	}
}

func TestGetWallets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK: true,
				Wallets: []domain.Wallet{
					{ID: 1, AccountNo: "KZT0000000001", Amount: 1},
					{ID: 2, AccountNo: "KZT0000000002", Amount: 232131},
					{ID: 3, AccountNo: "KZT0000000003", Amount: 0},
				},
			},
		)
	}))
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	wallets, err := api.GetWallets("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 3 {
		t.Errorf("Expecting %d wallets, got %d", 3, len(wallets))
	}
	ts.Close()
	// no transactions
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK:      true,
				Wallets: []domain.Wallet{},
			},
		)
	}))
	defer ts.Close()
	api = &WalletAPIInterface{ts.URL, newClient(ts.URL[7:])}
	wallets, err = api.GetWallets("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 0 {
		t.Errorf("Expecting 0 wallets, got %d", len(wallets))
	}
}

func TestGetWalletsErr(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte{})
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	wallets, err := api.GetWallets("", "")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(wallets) != 0 {
		t.Errorf("Expecting 0 wallets, got %d", len(wallets))
	}
	// no response at all
	api = &WalletAPIInterface{host: "nonexistent.com", client: newClient("nonexistent.com")}
	wallets, err = api.GetWallets("", "")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(wallets) != 0 {
		t.Errorf("Expecting 0 wallets, got %d", len(wallets))
	}
}

func TestAddWallet(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK:      true,
				Message: "Success",
			},
		)
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	message, err := api.AddWallet("")
	if err != nil {
		t.Fatal(err)
	}
	if len(message) == 0 {
		t.Errorf("Expecting a message, got none")
	}
}

func TestAddWalletErr(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte{})
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	message, err := api.AddWallet("")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(message) != 0 {
		t.Errorf("Expecting no message, got %s", message)
	}
	// no response at all
	api = &WalletAPIInterface{host: "nonexistent.com", client: newClient("nonexistent.com")}
	message, err = api.AddWallet("")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(message) != 0 {
		t.Errorf("Expecting no message, got %s", message)
	}
}

func TestGetWalletList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK: true,
				WalletList: []string{
					"KZT0000000001", "KZT0000000010", "KZT0000000001", "KZT0000000010", "KZT0000000001", "KZT0000000010", "KZT0000000001", "KZT0000000010",
				},
			},
		)
	}))
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	wallets, err := api.GetWalletList("")
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 8 {
		t.Errorf("Expecting %d wallets, got %d", 8, len(wallets))
	}
	ts.Close()
	// no transactions
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK:         true,
				WalletList: nil,
			},
		)
	}))
	defer ts.Close()
	api = &WalletAPIInterface{ts.URL, newClient(ts.URL[7:])}
	wallets, err = api.GetWalletList("")
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 0 {
		t.Errorf("Expecting 0 wallets, got %d", len(wallets))
	}
}

func TestGetWalletListErr(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte{})
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	wallets, err := api.GetWalletList("")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(wallets) != 0 {
		t.Errorf("Expecting 0 wallets, got %d", len(wallets))
	}
	// no response at all
	api = &WalletAPIInterface{host: "nonexistent.com", client: newClient("nonexistent.com")}
	wallets, err = api.GetWalletList("")
	if err == nil {
		t.Error("Expecting an error, got none")
	}
	if len(wallets) != 0 {
		t.Errorf("Expecting 0 wallets, got %d", len(wallets))
	}
}

func TestTopUp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK:      true,
				Message: "200",
			},
		)
	}))
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	res, _, err := api.TopUp("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Errorf("Expecting a message, got none")
	}
	ts.Close()
}

func TestTopUpErr(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte{})
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	res, _, err := api.TopUp("", "", "", "")
	if err != nil {
		t.Error("Expecting no error, got", err)
	}
	if len(res) != 0 {
		t.Errorf("Expecting no messages, got %s", res)
	}
	// no response at all
	api = &WalletAPIInterface{host: "nonexistent.com", client: newClient("nonexistent.com")}
	res, _, err = api.TopUp("", "", "", "")
	if err == nil {
		t.Error("Expecting no error, got", err)
	}
	if len(res) != 0 {
		t.Errorf("Expecting no messages, got %s", res)
	}
}

func TestTransfer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			domain.Response{
				OK:      true,
				Message: "Success",
			},
		)
	}))
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	res, _, err := api.Transfer("", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Errorf("Expecting a message, got none")
	}
	ts.Close()
}

func TestTransferErr(t *testing.T) {
	// incorrect response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte{})
	}))
	defer ts.Close()
	api := &WalletAPIInterface{host: ts.URL, client: newClient(ts.URL[7:])}
	res, _, err := api.Transfer("", "", "", "", "")
	if err != nil {
		t.Error("Expecting no error, got", err)
	}
	if len(res) != 0 {
		t.Errorf("Expecting no messages, got %s", res)
	}
	// no response at all
	api = &WalletAPIInterface{host: "nonexistent.com", client: newClient("nonexistent.com")}
	res, _, err = api.Transfer("", "", "", "", "")
	if err == nil {
		t.Error("Expecting no error, got", err)
	}
	if len(res) != 0 {
		t.Errorf("Expecting no messages, got %s", res)
	}
}
