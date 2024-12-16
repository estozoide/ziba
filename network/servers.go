package network

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
	"ziba/core"
	"ziba/store"
)

//
// SETUP (1/6)
//

// New.
func (s *SetupServer) New(store *store.BankStore) *SetupServer {
	s.port = setupPort
	s.store = store
	return s
}

// Start.
func (s *SetupServer) Start() error {
	// Start listening.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to start Setup server: %v", err)
		return err
	}

	log.Printf("Setup server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *SetupServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Setup]")

	// Close connection when finished.
	defer conn.Close()

	// Grab certificate file.
	directory, err := store.GetZibaDir()
	if err != nil {
		log.Fatalf("failed to retrieve Ziba directory: %v", err)
		return
	}
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", s.store.Name))
	file, err := os.Open(certPath)
	if err != nil {
		log.Fatalf("failed to open certificate file: %v", err)
		return
	}
	defer file.Close()

	// encoder := gob.NewEncoder(conn)
	writer := bufio.NewWriter(conn)

	// SEND name.
	bankName := s.store.Name
	if _, err := writer.WriteString(bankName + "\n"); err != nil {
		log.Fatalf("failed to encode Bank's name message: %v", err)
		return
	}

	// SEND file.
	_, err = io.Copy(writer, file)
	if err != nil {
		log.Fatalf("failed to send certificate file message: %v", err)
		return
	}

	// Flush writer.
	if err := writer.Flush(); err != nil {
		log.Fatalf("failed to flush connection: %v", err)
		return
	}

	// Info message.
	log.Print("Finished serving client [Setup]")
}

//
// ACCGEN (2/6)
//

// New.
func (s *AccgenServer) New(store *store.BankStore, config *tls.Config) *AccgenServer {
	s.port = accgenPort
	s.store = store
	s.config = config
	return s
}

// Start.
func (s *AccgenServer) Start() error {
	// Start listening.
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.port), s.config)
	if err != nil {
		log.Fatalf("failed to start Accgen server: %v", err)
		return err
	}

	log.Printf("Accgen server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *AccgenServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Accgen]")

	// Close connection when finished.
	defer conn.Close()

	// Read Bank.
	bank, err := s.store.ReadBank()
	if err != nil {
		log.Fatalf("failed to read Bank from database: %v", err)
		return
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// SEND BankProfile to client.
	bankProfile := bank.Profile()
	if err := encoder.Encode(*bankProfile); err != nil {
		log.Fatalf("failed to encode BankProfile message: %v", err)
		return
	}

	// RECV ClientProfile from client.
	var client core.ClientProfile
	if err := decoder.Decode(&client); err != nil {
		log.Fatalf("failed to decode ClientProfile message: %v", err)
		return
	}

	// Read ClientInfo from database. (Check if already in database)
	clientInfo, err := s.store.ReadClientInfo(&client)
	if clientInfo != nil {
		log.Fatalf("== ALERT: client already exists: %v", err)
		return
	} else if err != nil && err != sql.ErrNoRows {
		log.Fatalf("failed to read ClientInfo from database: %v", err)
		return
	}

	// Create client account.
	clientInfo, err = bank.NewClient(&client)
	if err != nil {
		log.Fatalf("failed to create client account: %v", err)
		return
	}

	// Write ClientInfo.
	if err := s.store.WriteClientInfo(clientInfo); err != nil {
		log.Fatalf("failed to write ClientInfo into database: %v", err)
		return
	}

	// SEND credentials to client.
	credentials := struct {
		Credential *big.Int
		Contract   *big.Int
	}{
		Credential: clientInfo.Credential,
		Contract:   clientInfo.Contract,
	}
	if err := encoder.Encode(credentials); err != nil {
		log.Fatalf("failed to encode ClientInfo message: %v", err)
		return
	}

	// Info message.
	log.Printf("ClientInfo: %s", clientInfo)
	log.Print("Finished serving client [Accgen]")
}

//
// WITHDRAWAL (3/6)
//

// New.
func (s *WithdrawalServer) New(store *store.BankStore, config *tls.Config) *WithdrawalServer {
	s.port = withdrawalPort
	s.store = store
	s.config = config
	return s
}

// Start.
func (s *WithdrawalServer) Start() error {
	// Start listening.
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.port), s.config)
	if err != nil {
		log.Fatalf("failed to start Withdrawal server: %v", err)
		return err
	}

	log.Printf("Withdrawal server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *WithdrawalServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Withdrawal]")

	// Close connection when finished.
	defer conn.Close()

	// Read Bank.
	bank, err := s.store.ReadBank()
	if err != nil {
		log.Fatalf("failed to read Bank from database: %v", err)
		return
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// RECV client profile.
	var client core.ClientProfile
	if err := decoder.Decode(&client); err != nil {
		log.Fatalf("failed to decode ClientProfile message: %v", err)
		return
	}

	// RECV coin request.
	var request struct {
		ALower *big.Int
		C      *big.Int
	}
	if err := decoder.Decode(&request); err != nil {
		log.Fatalf("failed to decode Withdrawal request message: %v", err)
		return
	}

	// Read ClientInfo from database. (Check that exists)
	clientInfo, err := s.store.ReadClientInfo(&client)
	if clientInfo == nil {
		log.Fatalf("== ALERT: client does not exist in database: %v", err)
		return
	} else if err != nil && err != sql.ErrNoRows {
		log.Fatalf("failed to read ClientInfo from database: %v", err)
		return
	}

	// Grab client's balance.
	balance, err := s.store.ReadClientBalance(&client)
	if err != nil {
		log.Fatalf("failed to read client's balance from database: %v", err)
		return
	}

	// Check if balance is sufficient.
	if balance < 1 {
		log.Print("Insufficient funds")
		return
	}

	// Update client's balance.
	err = s.store.UpdateClientBalance(&client, balance-1)
	if err != nil {
		log.Fatalf("failed to update client's balance into database: %v", err)
		return
	}

	// Compute coin response.
	Expiration, A1, C1 := bank.NewCoinResponse(clientInfo, request.ALower, request.C)

	// Craft response.
	response := struct {
		Expiration time.Time
		A1         *big.Int
		C1         *big.Int
	}{
		Expiration: Expiration,
		A1:         A1,
		C1:         C1,
	}

	// SEND response.
	if err := encoder.Encode(response); err != nil {
		log.Fatalf("failed to encode Withdrawal response message: %v", err)
		return
	}

	// Info message.
	log.Print("Finished serving client [Withdrawal]")
}

//
// PAYMENT (4/6)
//

// New.
func (s *PaymentServer) New(store *store.ClientStore, config *tls.Config) *PaymentServer {
	s.port = paymentPort
	s.store = store
	s.config = config
	return s
}

// Start.
func (s *PaymentServer) Start() error {
	// Start listening.
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.port), s.config)
	if err != nil {
		log.Fatalf("failed to start Payment server: %v", err)
		return err
	}

	log.Printf("Payment server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *PaymentServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Payment]")

	// Close connection when finished.
	defer conn.Close()

	// Read Client.
	client, err := s.store.ReadClient()
	if err != nil {
		log.Fatalf("failed to read Client from database: %v", err)
		return
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// RECV CoinProfile.
	var coin core.CoinProfile
	if err := decoder.Decode(&coin); err != nil {
		log.Fatalf("failed to decode CoinProfile message: %v", err)
		return
	}

	// Verify coin properties.
	if valid := coin.VerifyProperties(&client.Bank); !valid {
		log.Print("invalid Coin")
		return
	}

	// Stamp coin.
	msg := coin.Stamp(&client.Bank, client.Profile())

	// SEND Elgamal's msg.
	if err := encoder.Encode(msg); err != nil {
		log.Fatalf("failed to encode Elgamal's msg message: %v", err)
		return
	}

	// RECV Elgamal's second.
	var second *big.Int
	if err := decoder.Decode(&second); err != nil {
		log.Fatalf("failed to decode Elgamal's second message: %v", err)
		return
	}

	// Verify Elgamal signature.
	if valid := coin.VerifyElgamal(&client.Bank, second); !valid {
		log.Fatalf("invalid Elgamal's signature")
		return
	}

	// SEND acceptance.
	accept := true
	encoder.Encode(accept)

	// Write coin.
	newCoin := core.Coin{
		Random: core.CoinRandom{},
		Elgamal: core.CoinElgamal{
			Pub:    coin.Pub,
			First:  coin.First,
			Second: second,
			Msg:    msg,
		},
		Params: core.CoinParams{
			A:          coin.A,
			A2:         coin.A2,
			R:          coin.R,
			Expiration: coin.Expiration,
		},
	}
	if err := s.store.WriteCoin(&newCoin, store.Operation_Payment); err != nil {
		log.Fatalf("failed to write Coin into database: %v", err)
		return
	}

	// Info message.
	log.Print("Finished serving client [Payment]")
}

//
// DEPOSIT (5/6)
//

// New.
func (s *DepositServer) New(store *store.BankStore, config *tls.Config) *DepositServer {
	s.port = depositPort
	s.store = store
	s.config = config
	return s
}

// Start.
func (s *DepositServer) Start() error {
	// Start listening.
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.port), s.config)
	if err != nil {
		log.Fatalf("failed to start Deposit server: %v", err)
		return err
	}

	log.Printf("Deposit server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *DepositServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Deposit]")

	// Close connection when finished.
	defer conn.Close()

	// Read Bank.
	bank, err := s.store.ReadBank()
	if err != nil {
		log.Fatalf("failed to read Bank from database: %v", err)
		return
	}
	bankProfile := bank.Profile()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// RECV client profile.
	var client core.ClientProfile
	if err := decoder.Decode(&client); err != nil {
		log.Fatalf("failed to decode ClientProfile message: %v", err)
		return
	}

	// Read ClientInfo from database. (Check that exists)
	clientInfo, err := s.store.ReadClientInfo(&client)
	if clientInfo == nil {
		log.Fatalf("== ALERT: client does not exist in database: %v", err)
		return
	} else if err != nil && err != sql.ErrNoRows {
		log.Fatalf("failed to read ClientInfo from database: %v", err)
		return
	}

	// RECV coin profile.
	var coin core.CoinProfile
	if err := decoder.Decode(&coin); err != nil {
		log.Fatalf("failed to decode CoinProfile message: %v", err)
	}

	// Verify coin properties.
	if valid := coin.VerifyProperties(bankProfile); !valid {
		log.Fatalf("invalid coin")
		return
	}

	// Read coin profile from database. (Check if already in database)
	err = s.store.ReadCoinProfile(&coin)
	if err == sql.ErrNoRows {
		// all good
	} else if err != nil {
		log.Fatalf("failed to read CoinProfile from database: %v", err)
		return
	}

	// Write coin profile into database.
	if err := s.store.WriteCoinProfile(&coin, store.Operation_Deposit, &client); err != nil {
		log.Fatalf("failed to write CoinProfile into database: %v", err)
		return
	}

	// Grab client's balance.
	balance, err := s.store.ReadClientBalance(&client)
	if err != nil {
		log.Fatalf("failed to read client's balance from database: %v", err)
		return
	}

	// Update client's balance.
	err = s.store.UpdateClientBalance(&client, balance+1)
	if err != nil {
		log.Fatalf("failed to update client's balance into database: %v", err)
		return
	}

	// Craft response.
	accept := true

	// SEND response.
	if err := encoder.Encode(accept); err != nil {
		log.Fatalf("failed to encode Response message: %v", err)
		return
	}

	// Info message.
	log.Print("Finished serving client [Deposit]")
}

//
// EXCHANGE (6/6)
//

// New.
func (s *ExchangeServer) New(store *store.BankStore, config *tls.Config) *ExchangeServer {
	s.port = exchangePort
	s.store = store
	s.config = config
	return s
}

// Start.
func (s *ExchangeServer) Start() error {
	// Start listening.
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.port), s.config)
	if err != nil {
		log.Fatalf("failed to start Exchange server: %v", err)
		return err
	}

	log.Printf("Exchange server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *ExchangeServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Exchange]")

	// Close connection when finished.
	defer conn.Close()

	// Read Bank.
	bank, err := s.store.ReadBank()
	if err != nil {
		log.Fatalf("failed to read Bank from database: %v", err)
		return
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// RECV client profile.
	var client core.ClientProfile
	if err := decoder.Decode(&client); err != nil {
		log.Fatalf("failed to decode ClientProfile message: %v", err)
		return
	}

	// RECV coin profile.
	var coin core.CoinProfile
	if err := decoder.Decode(&coin); err != nil {
		log.Fatalf("failed to decode CoinProfile message: %v", err)
		return
	}

	// RECV coin request.
	var request struct {
		ALower *big.Int
		C      *big.Int
	}
	if err := decoder.Decode(&request); err != nil {
		log.Fatalf("failed to decode Exchange request message: %v", err)
		return
	}

	// Read ClientInfo from database. (Check that exists)
	clientInfo, err := s.store.ReadClientInfo(&client)
	if clientInfo == nil {
		log.Fatalf("== ALERT: client does not exist in database: %v", err)
		return
	} else if err != nil && err != sql.ErrNoRows {
		log.Fatalf("failed to read ClientInfo from database: %v", err)
		return
	}

	// Verify coin.
	if valid := coin.VerifyProperties(bank.Profile()); !valid {
		log.Fatalf("invalid coin")
		return
	}

	// Read coin profile from database. (Check if already in database)
	err = s.store.ReadCoinProfile(&coin)
	if err == sql.ErrNoRows {
		// all good
	} else if err != nil {
		log.Fatalf("failed to read CoinProfile from database: %v", err)
		return
	}

	// Write coin profile into database.
	if err := s.store.WriteCoinProfile(&coin, store.Operation_Exchange, &client); err != nil {
		log.Fatalf("failed to write CoinProfile into database: %v", err)
		return
	}

	// Check Expiration date of coin.
	now := time.Now()
	if valid := coin.Expiration.After(now); valid {
		duration := coin.Expiration.Sub(now)
		months := int(duration.Hours()/24/30) % 12
		days := int(duration.Hours()/24) % 30
		hours := int(duration.Hours()) % 24
		log.Printf("Coin is still valid for %d months, %d days, %d hours", months, days, hours)
		// return
	}

	// Compute coin response.
	Expiration, A1, C1 := bank.NewCoinResponse(clientInfo, request.ALower, request.C)

	// Craft response.
	response := struct {
		Expiration time.Time
		A1         *big.Int
		C1         *big.Int
	}{
		Expiration: Expiration,
		A1:         A1,
		C1:         C1,
	}

	// SEND coin response.
	if err := encoder.Encode(response); err != nil {
		log.Fatalf("failed to encode Exchange response message: %v", err)
		return
	}

	// Info message.
	log.Print("Finished serving client [Exchange]")
}

//
// GET
//

// New.
func (s *GetServer) New(filepath string) *GetServer {
	s.port = getPort
	s.filepath = filepath
	return s
}

// Start.
func (s *GetServer) Start() error {
	// Start listening.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to start Get server: %v", err)
		return err
	}

	log.Printf("Get server listening on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

// handleClient.
func (s *GetServer) handleClient(conn net.Conn) {
	// Info message.
	log.Print("Serving client [Get]")

	// Close connection when finished.
	defer conn.Close()

	// Grab file.
	file, err := os.Open(s.filepath)
	if err != nil {
		log.Fatalf("failed to open file %s: %v", s.filepath, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(conn)

	// SEND file.
	_, err = io.Copy(writer, file)
	if err != nil {
		log.Fatalf("failed to send file message: %v", err)
		return
	}

	// Flush writer.
	if err := writer.Flush(); err != nil {
		log.Fatalf("failed to flush connection: %v", err)
		return
	}

	// Info message.
	log.Print("Finished serving client [Get]")
}
