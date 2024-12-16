package store_test

import (
	"log"
	"path/filepath"
	"testing"
	"ziba/core"
	"ziba/store"
)

var (
	zibaDir    string
	bank       *core.Bank
	client     *core.Client
	clientInfo *core.ClientInfo
	coin       *core.Coin
)

func TestMain(m *testing.M) {
	// Get Ziba directory.
	zibaDir, _ = store.GetZibaDir()

	// Load scheme parameters.
	scheme := core.Params

	// SETUP

	// Create bank.
	bank = new(core.Bank).New(scheme)
	bankProfile := bank.Profile()

	// ACCGEN

	// Create client.
	client = new(core.Client).New(bankProfile)
	clientProfile := client.Profile()

	// Create client account.
	clientInfo, _ = bank.NewClient(clientProfile)
	client.SetCredentials(clientInfo.Credential, clientInfo.Contract)

	// WITHDRAWAL

	// Create coin request.
	coin = client.NewCoinRequest()

	// Create coin response.
	Expiration, A1, C1 := bank.NewCoinResponse(clientInfo, coin.Params.ALower, coin.Params.C)

	// Build final coin.
	client.FinishCoin(coin, Expiration, A1, C1)

	// Run tests.
	m.Run()
}

const (
	identity = "main"
	bankName = "BanCoco"
)

func TestBankStore(t *testing.T) {
	// Grab database path.
	dbPath := filepath.Join(zibaDir, "bank.db")

	// New.
	bankStore, err := new(store.BankStore).New(dbPath, identity)
	if err != nil {
		t.Fatal(err)
	}

	// WriteBank.
	err = bankStore.WriteBank(bank, bankName)
	if err != nil {
		t.Fatal(err)
	}

	// ReadBank.
	bank, err = bankStore.ReadBank()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(bank)

	// WriteClientInfo.
	err = bankStore.WriteClientInfo(clientInfo)
	if err != nil {
		t.Fatal(err)
	}

	// ReadClientInfo.
	clientInfo, err = bankStore.ReadClientInfo(client.Profile())
	if err == store.ErrExistingClient {
		t.Log("client already exists")
	} else if err != nil {
		t.Fatal(err)
	}
	t.Log(clientInfo)

	// WriteCoinProfile.
	err = bankStore.WriteCoinProfile(coin.Profile(), store.Operation_Deposit, &clientInfo.Profile)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(coin.Profile())

	// ReadCoinProfile.
	err = bankStore.ReadCoinProfile(coin.Profile())
	if err == store.ErrExistingCoin {
		t.Log("coin already exists")
	} else if err != nil {
		t.Fatal(err)
	}
}

func TestClientStore(t *testing.T) {
	// Grab database path.
	dbPath := filepath.Join(zibaDir, "client.db")

	// New.
	clientStore, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	clientStore.BankName = bankName

	// WriteClient.
	err = clientStore.WriteClient(client)
	if err != nil {
		t.Fatal(err)
	}

	// ReadClient.
	client, err = clientStore.ReadClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(client)

	// WriteCoin.
	err = clientStore.WriteCoin(coin, store.Operation_Withdrawal)
	if err != nil {
		t.Fatal(err)
	}

	// ReadCoins.
	coins, err := clientStore.ReadCoins()
	if err != nil {
		t.Fatal(err)
	}
	for _, coin := range coins {
		t.Log(coin)
	}
	t.Logf("total coins: %d", len(coins))
}

func TestStoreCoins(t *testing.T) {
	directory, _ := store.GetZibaDir()
	dbPath := filepath.Join(directory, "agus.db")
	store, _ := new(store.ClientStore).New(dbPath)
	store.BankName = "bancoco"
	client, _ := store.ReadClient()
	coins, _ := store.ReadCoins()
	for _, coin := range coins {
		valid := coin.Profile().VerifyProperties(&client.Bank)
		log.Printf("%v", valid)
	}
}
