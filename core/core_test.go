package core_test

import (
	"testing"
	"ziba/core"
)

func TestCore(t *testing.T) {
	// Get scheme parameters.
	scheme := core.Params

	// SETUP

	// Create bank.
	bank := new(core.Bank).New(scheme)
	bankProfile := bank.Profile()
	t.Log(bank)
	t.Log(bankProfile)

	// ACCOUNT GENERATION

	// Create client.
	client := new(core.Client).New(bankProfile)
	clientProfile := client.Profile()
	t.Log(client)
	t.Log(clientProfile)

	// Create client account.
	clientInfo, err := bank.NewClient(clientProfile)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(clientInfo)

	// Add credentials.
	client.SetCredentials(clientInfo.Credential, clientInfo.Contract)
	t.Log(client)

	// WITHDRAWAL

	// Create request.
	coin := client.NewCoinRequest()
	t.Log(coin)

	// Create response.
	Expiration, A1, C1 := bank.NewCoinResponse(clientInfo, coin.Params.ALower, coin.Params.C)

	// Build final coin.
	client.FinishCoin(coin, Expiration, A1, C1)
	t.Log(coin)

	coinProfile := coin.Profile()
	t.Log(coinProfile)

	// PAYMENT

	valid := coinProfile.VerifyProperties(bankProfile)
	if !valid {
		t.Fatalf("invalid")
	}
	t.Log("Valid Coin properties")

	msg := coinProfile.Stamp(bankProfile, clientProfile)
	t.Log(coinProfile)

	second := client.SignCoin(coin, msg)
	t.Log(coin)

	valid = coinProfile.VerifyElgamal(bankProfile, second)
	if !valid {
		t.Fatalf("invalid")
	}
	t.Log("Valid Elgamal's signature")

}
