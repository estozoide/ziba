package network

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"ziba/core"
	"ziba/store"
)

//
// SETUP (1/6)
//

// New.
func (c *SetupClient) New(serverAddr string, store *store.ClientStore) *SetupClient {
	c.serverAddr = serverAddr
	c.store = store
	return c
}

// Execute.
func (c *SetupClient) Execute() error {
	// Connect to server.
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, setupPort))
	if err != nil {
		log.Fatalf("failed to connect to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Printf("Connected to Setup server")

	// Create a file to copy into the certificate.
	directory, err := store.GetZibaDir()
	if err != nil {
		log.Fatalf("failed to retrieve Ziba directory: %v", err)
		return err
	}
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", c.serverAddr))
	certFile, err := os.Create(certPath)
	if err != nil {
		log.Printf("failed to create certificate file: %v", err)
		return err
	}
	defer certFile.Close()

	// decoder := gob.NewDecoder(conn)
	reader := bufio.NewReader(conn)

	// RECV name.
	bankName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to decode Bank's name message: %v", err)
		return err
	}
	c.store.BankName = strings.TrimSpace(bankName)
	log.Printf("\n\n  Hello,\n  Welcome to %s\n\n", bankName)

	// RECV file.
	_, err = io.Copy(certFile, reader)
	if err != nil {
		log.Fatalf("failed to read certificate file message: %v", err)
		return err
	}

	// Info message.
	log.Printf("Certificate downloaded")

	return nil
}

//
// ACCGEN (2/6)
//

// New.
func (c *AccgenClient) New(serverAddr string, store *store.ClientStore, config *tls.Config) *AccgenClient {
	c.serverAddr = serverAddr
	c.store = store
	c.config = config
	return c
}

// Execute.
func (c *AccgenClient) Execute() error {
	// Connect to server.
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, accgenPort), c.config)
	if err != nil {
		log.Fatalf("failed to connect to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Print("Connected to Accgen server")

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// RECV BankProfile from server.
	var bankProfile core.BankProfile
	if err := decoder.Decode(&bankProfile); err != nil {
		log.Fatalf("failed to decode BankProfile message: %v", err)
		return err
	}

	// Create Client.
	client := new(core.Client).New(&bankProfile)
	clientProfile := client.Profile()

	// SEND ClientProfile to server.
	if err := encoder.Encode(*clientProfile); err != nil {
		log.Fatalf("failed to encode ClientProfile message: %v", err)
		return err
	}

	// RECV credentials from server.
	var credentials struct {
		Credential *big.Int
		Contract   *big.Int
	}
	if err := decoder.Decode(&credentials); err != nil {
		log.Fatalf("failed to decode ClientInfo message: %v", err)
		return err
	}

	// Add credentials.
	client.SetCredentials(credentials.Credential, credentials.Contract)

	// Write Client into database.
	if err := c.store.WriteClient(client); err != nil {
		log.Fatalf("failed to write Client into database: %v", err)
		return err
	}

	// Info message.
	log.Printf("Client: %s", client)
	log.Printf("Account Generation Success!")

	return nil
}

//
// WITHDRAWAL (3/6)
//

// New.
func (c *WithdrawalClient) New(serverAddr string, store *store.ClientStore, config *tls.Config) *WithdrawalClient {
	c.serverAddr = serverAddr
	c.store = store
	c.config = config
	return c
}

// Execute.
func (c *WithdrawalClient) Execute() error {
	// Connect to server.
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, withdrawalPort), c.config)
	if err != nil {
		log.Fatalf("failed to connect to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Print("Connected to Withdrawal server")

	// Read Client.
	client, err := c.store.ReadClient()
	if err != nil {
		log.Fatalf("failed to read Client from database: %v", err)
		return err
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// Fake Client.
	// client2 := new(core.Client).New(&client.Bank)
	// client2Profile := client2.Profile()

	// SEND client profile.
	clientProfile := client.Profile()
	if err := encoder.Encode(*clientProfile); err != nil {
		log.Fatalf("failed to encode ClientProfile message: %v", err)
		return err
	}

	// Compute coin request.
	coin := client.NewCoinRequest()

	// Craft request.
	request := struct {
		ALower *big.Int
		C      *big.Int
	}{
		ALower: coin.Params.ALower,
		C:      coin.Params.C,
	}

	// SEND coin request.
	if err := encoder.Encode(request); err != nil {
		log.Fatalf("failed to encode Withdrawal request message: %v", err)
		return err
	}

	// RECV coin response.
	var response struct {
		Expiration time.Time
		A1         *big.Int
		C1         *big.Int
	}
	if err := decoder.Decode(&response); err != nil {
		log.Fatalf("failed to decode Withdrawal response message: %v", err)
		return err
	}

	// Finish the coin using response.
	client.FinishCoin(coin, response.Expiration, response.A1, response.C1)

	// Write coin.
	if err := c.store.WriteCoin(coin, store.Operation_Withdrawal); err != nil {
		log.Fatalf("failed to write Coin into database: %v", err)
		return err
	}

	// Info mesage.
	log.Printf("Coin: %s", coin)
	log.Printf("Withdrawal Success!")

	return nil
}

//
// PAYMENT (4/6)
//

// New.
func (c *PaymentClient) New(serverAddr string, store *store.ClientStore, config *tls.Config) *PaymentClient {
	c.serverAddr = serverAddr
	c.store = store
	c.config = config
	return c
}

// Execute.
func (c *PaymentClient) Execute() error {
	// Connect to server.
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, paymentPort), c.config)
	if err != nil {
		log.Fatalf("failed to connect to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Print("Connected to Payment server")

	// Read Client.
	client, err := c.store.ReadClient()
	if err != nil {
		log.Fatalf("failed to read Client from database: %v", err)
		return err
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// Read coins.
	coins, err := c.store.ReadCoins()
	if err != nil {
		log.Fatalf("failed to read coins from database: %v", err)
		return err
	}

	// Check local balance.
	balance := len(coins)
	// log.Printf("Current balance: %d", balance)
	if balance < 1 {
		log.Printf("No coins on local storage")
		return nil
	}

	// Grab 1 coin.
	coin := coins[0]
	coinProfile := coin.Profile()

	// SEND CoinProfile.
	if err := encoder.Encode(*coinProfile); err != nil {
		log.Fatalf("failed to encode CoinProfile message: %v", err)
		return err
	}

	// RECV Elgamal's msg.
	var msg *big.Int
	if err := decoder.Decode(&msg); err != nil {
		log.Fatalf("failed to decode Elgamal's msg message: %v", err)
		return err
	}

	// Sign coin.
	second := client.SignCoin(&coin, msg)

	// SEND Elgamal's second.
	if err := encoder.Encode(second); err != nil {
		log.Fatalf("failed to encode Elgamal's second message: %v", err)
		return err
	}

	// RECV acceptance.
	var accept bool
	if err := decoder.Decode(&accept); err != nil {
		log.Fatalf("failed to decode acceptance message: %v", err)
		return err
	}

	// Delete Coin after payment.
	if accept {
		if err := c.store.DeleteCoin(&coin, store.Operation_Payment); err != nil {
			log.Fatalf("failed to delete coin from database: %v", err)
		}
	}

	// Info message.
	log.Printf("Current balance: %d", balance-1)
	log.Printf("Payment Success!")

	return nil
}

//
// DEPOSIT (5/6)
//

// New.
func (c *DepositClient) New(serverAddr string, store *store.ClientStore, config *tls.Config) *DepositClient {
	c.serverAddr = serverAddr
	c.store = store
	c.config = config
	return c
}

// Execute.
func (c *DepositClient) Execute() error {
	// Connect to server.
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, depositPort), c.config)
	if err != nil {
		log.Fatalf("failed to connect to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Print("Connected to Deposit server")

	// Read Client.
	client, err := c.store.ReadClient()
	if err != nil {
		log.Fatalf("failed to read Client from database: %v", err)
		return err
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// Read coins.
	coins, err := c.store.ReadCoins()
	if err != nil {
		log.Fatalf("failed to read coins from database: %v", err)
		return err
	}

	// Check local balance.
	balance := len(coins)
	if balance < 1 {
		log.Printf("No coins on local storage")
		return nil
	}

	// Grab 1 coin.
	coin := coins[0]
	coinProfile := coin.Profile()

	// SEND ClientProfile.
	clientProfile := client.Profile()
	if err := encoder.Encode(*clientProfile); err != nil {
		log.Fatalf("failed to encode ClientProfile message: %v", err)
		return err
	}

	// SEND CoinProfile.
	if err := encoder.Encode(*coinProfile); err != nil {
		log.Fatalf("failed to encode CoinProfile message: %v", err)
		return err
	}

	// RECV response.
	var accept bool
	if err := decoder.Decode(&accept); err != nil {
		log.Fatalf("failed to decode Deposit response message: %v", err)
		return err
	}

	// Delete Coin after deposit.
	if accept {
		if err := c.store.DeleteCoin(&coin, store.Operation_Deposit); err != nil {
			log.Fatalf("failed to delete coin from database: %v", err)
		}
	}

	// Info message.
	log.Printf("Balance: %d", balance-1)
	log.Printf("Deposit Success!")

	return nil
}

//
// EXCHANGE (6/6)
//

// New.
func (c *ExchangeClient) New(serverAddr string, store *store.ClientStore, config *tls.Config) *ExchangeClient {
	c.serverAddr = serverAddr
	c.store = store
	c.config = config
	return c
}

// Execute.
func (c *ExchangeClient) Execute() error {
	// Connect to server.
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, exchangePort), c.config)
	if err != nil {
		log.Fatalf("failed to connect to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Print("Connected to Exchange server")

	// Read Client.
	client, err := c.store.ReadClient()
	if err != nil {
		log.Fatalf("failed to read Client from database: %v", err)
		return err
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// Read coins.
	coins, err := c.store.ReadCoins()
	if err != nil {
		log.Fatalf("failed to read coins from database: %v", err)
		return err
	}

	// Check local balance.
	balance := len(coins)
	if balance < 1 {
		log.Printf("No coins on local storage")
		return nil
	}

	// Grab 1 coin.
	coin := coins[0]
	coinProfile := coin.Profile()

	// SEND client profile.
	clientProfile := client.Profile()
	if err := encoder.Encode(*clientProfile); err != nil {
		log.Fatalf("failed to encode ClientProfile message: %v", err)
		return err
	}

	// SEND CoinProfile.
	if err := encoder.Encode(*coinProfile); err != nil {
		log.Fatalf("failed to encode CoinProfile message: %v", err)
		return err
	}

	// Compute coin request.
	newCoin := client.NewCoinRequest()

	// Craft request.
	request := struct {
		ALower *big.Int
		C      *big.Int
	}{
		ALower: newCoin.Params.ALower,
		C:      newCoin.Params.C,
	}

	// SEND coin request.
	if err := encoder.Encode(request); err != nil {
		log.Fatalf("failed to encode Withdrawal request message: %v", err)
		return err
	}

	// RECV coin response.
	var response struct {
		Expiration time.Time
		A1         *big.Int
		C1         *big.Int
	}
	if err := decoder.Decode(&response); err != nil {
		log.Fatalf("failed to decode Withdrawal response message: %v", err)
		return err
	}

	// Finish the coin using response.
	client.FinishCoin(newCoin, response.Expiration, response.A1, response.C1)

	// Write coin.
	if err := c.store.WriteCoin(newCoin, store.Operation_Exchange); err != nil {
		log.Fatalf("failed to write Coin into database: %v", err)
		return err
	}

	// Delete previous coin.
	if err := c.store.DeleteCoin(&coin, store.Operation_Exchange); err != nil {
		log.Fatalf("failed to delete coin from database: %v", err)
	}

	// Info message.
	log.Printf("Coin: %s", newCoin)
	log.Printf("Exchange Success!")

	return nil
}

//
// GET
//

// New.
func (c *GetClient) New(serverAddr string) *GetClient {
	c.serverAddr = serverAddr
	return c
}

// Execute.
func (c *GetClient) Execute() error {
	// Connect to server.
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.serverAddr, getPort))
	if err != nil {
		log.Fatalf("failed to connecto to server at %s: %v", c.serverAddr, err)
		return err
	}
	defer conn.Close()

	// Info message.
	log.Printf("Connected to Get server")

	// Create file to copy into.
	directory, err := store.GetZibaDir()
	if err != nil {
		log.Fatalf("failed to retrieve Ziba directory: %v", err)
		return err
	}
	filepath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", c.serverAddr))
	file, err := os.Create(filepath)
	if err != nil {
		log.Printf("failed to create file: %v", err)
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(conn)

	// RECV file.
	_, err = io.Copy(file, reader)
	if err != nil {
		log.Fatalf("failed to read file message: %v", err)
		return err
	}

	// Info message.
	log.Printf("Get Success!")

	return nil
}
