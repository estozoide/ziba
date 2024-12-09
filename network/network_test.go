package network_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"ziba/core"
	"ziba/network"
	"ziba/store"
)

var (
	address   = "localhost"
	bankName  = "bancoco"
	userName  = "carlos"
	userName2 = "usuario"
)

func TestInit(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		log.Fatal(err)
	}

	// Create BankStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", bankName))
	store, err := new(store.BankStore).New(dbPath, "main")
	if err != nil {
		log.Fatal(err)
	}

	// Create Bank.
	bank := new(core.Bank).New(core.Params)

	// Write Bank into store.
	store.WriteBank(bank, bankName)

	// Create key and certificate for Bank.
	err = network.CreateCertificate(directory, bankName)
	if err != nil {
		log.Fatal(err)
	}

	// Make a copy of bank's certificate.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", bankName))
	certCopyPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))

	certFile, err := os.Open(certPath)
	if err != nil {
		t.Fatal(err)
	}
	defer certFile.Close()

	certCopyFile, err := os.Create(certCopyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer certCopyFile.Close()

	_, err = io.Copy(certCopyFile, certFile)
	if err != nil {
		t.Fatal(err)
	}

	// Create key and certificate for User 1.
	err = network.CreateCertificate(directory, userName)
	if err != nil {
		log.Fatal(err)
	}

	// Make a copy of user's certificate.
	certPath = filepath.Join(directory, fmt.Sprintf("%s_cert.pem", userName))
	certCopyPath = filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", userName))

	certFile, err = os.Open(certPath)
	if err != nil {
		t.Fatal(err)
	}
	defer certFile.Close()

	certCopyFile, err = os.Create(certCopyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer certCopyFile.Close()

	_, err = io.Copy(certCopyFile, certFile)
	if err != nil {
		t.Fatal(err)
	}
}

// ***********
// SETUP (1/6)
// ***********

func TestSetupServer(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create BankStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", bankName))
	store, err := new(store.BankStore).New(dbPath, "main")
	if err != nil {
		t.Fatal(err)
	}

	// New.
	server := new(network.SetupServer).New(store)

	// Start.
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
}

func TestSetupClient(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// New.
	client := new(network.SetupClient).New(address, store)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

// ************
// ACCGEN (2/6)
// ************

func TestAccgenServer(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create BankStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", bankName))
	store, err := new(store.BankStore).New(dbPath, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Load TLS server configuration.
	keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", bankName))
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", bankName))
	config, err := network.GetServerTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("failed to grab TLS server configuration: %v", err)
	}

	// New.
	server := new(network.AccgenServer).New(store, config)

	// Start.
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
}

func TestAccgenClient(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.AccgenClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAccgenClient2(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName2))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.AccgenClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

// ****************
// WITHDRAWAL (3/6)
// ****************

func TestWithdrawalServer(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create BankStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", bankName))
	store, err := new(store.BankStore).New(dbPath, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Load TLS server configuration.
	keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", bankName))
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", bankName))
	config, err := network.GetServerTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("failed to grab TLS server configuration: %v", err)
	}

	// New.
	server := new(network.WithdrawalServer).New(store, config)

	// Start.
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
}

func TestWithdrawalClient(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.WithdrawalClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestWithdrawalClient2(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName2))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.WithdrawalClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

// *************
// PAYMENT (4/6)
// *************

func TestPaymentServer(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS server configuration.
	keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", userName))
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", userName))
	config, err := network.GetServerTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("failed to grab TLS server configuration: %v", err)
	}

	// New.
	server := new(network.PaymentServer).New(store, config)

	// Start.
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
}

func TestPaymentClient(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName2))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", userName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.PaymentClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

// *************
// DEPOSIT (5/6)
// *************

func TestDepositServer(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create BankStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", bankName))
	store, err := new(store.BankStore).New(dbPath, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Load TLS server configuration.
	keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", bankName))
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", bankName))
	config, err := network.GetServerTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("failed to grab TLS server configuration: %v", err)
	}

	// New.
	server := new(network.DepositServer).New(store, config)

	// Start.
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
}

func TestDepositClient(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.DepositClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}

// **************
// EXCHANGE (6/6)
// **************

func TestExchangeServer(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create BankStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", bankName))
	store, err := new(store.BankStore).New(dbPath, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Load TLS server configuration.
	keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", bankName))
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", bankName))
	config, err := network.GetServerTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("failed to grab TLS server configuration: %v", err)
	}

	// New.
	server := new(network.ExchangeServer).New(store, config)

	// Start.
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
}

func TestExchangeClient(t *testing.T) {
	// Get Ziba directory.
	directory, err := store.GetZibaDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create ClientStore.
	dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", userName))
	store, err := new(store.ClientStore).New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	store.BankName = bankName

	// Load TLS client configuration.
	certPath := filepath.Join(directory, fmt.Sprintf("%s_cert_cpy.pem", bankName))
	config, err := network.GetClientTLSConfig(certPath)
	if err != nil {
		t.Fatalf("failed to grab TLS client configuration: %v", err)
	}

	// New.
	client := new(network.ExchangeClient).New(address, store, config)

	// Execute.
	if err := client.Execute(); err != nil {
		t.Fatal(err)
	}
}
