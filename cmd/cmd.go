package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"ziba/core"
	"ziba/network"
	"ziba/store"

	"github.com/spf13/cobra"
)

// flags
var (
	flags struct {
		address  string
		bank     string
		identity string
		user     string
		inspect  bool
	}
)

// ziba
var ziba = &cobra.Command{
	Use:   "ziba command",
	Short: "A cryptographic-based CLI payment application.",
}

// user
var user = &cobra.Command{
	Use:   "user sub-command",
	Short: "Perform user operations.",
}

// user init
var userInit = &cobra.Command{
	Use:   "init --user USER",
	Short: "Create a new user named USER.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve Ziba directory: %v", err)
		}

		// Create local database.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		new(store.ClientStore).New(dbPath)

		// Create certificates.
		network.CreateCertificate(directory, flags.user)
	},
}

// user request-account
var accgen = &cobra.Command{
	Use:   "accgen --user USER --server SERVER",
	Short: "Request a client account for USER at SERVER.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}

		if len(flags.address) == 0 {
			return fmt.Errorf("required \"server\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		// Execute SetupClient.
		setupClient := new(network.SetupClient).New(flags.address, store)
		if err := setupClient.Execute(); err != nil {
			log.Fatal(err)
		}

		// Load TLS client configuration.
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.address))
		config, err := network.GetClientTLSConfig(certPath)
		if err != nil {
			log.Fatalf("failed to load certificate (client): %v", err)
		}

		// Execute AccgenClient.
		client := new(network.AccgenClient).New(flags.address, store, config)
		if err := client.Execute(); err != nil {
			log.Fatal(err)
		}
	},
}

// user withdraw
var withdraw = &cobra.Command{
	Use:   "withdraw --user USER --server SERVER",
	Short: "Withdraw 1 coin from USER's client account at SERVER.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}

		if len(flags.address) == 0 {
			return fmt.Errorf("required \"server\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		// Execute SetupClient.
		setupClient := new(network.SetupClient).New(flags.address, store)
		if err := setupClient.Execute(); err != nil {
			log.Fatal(err)
		}

		// Load TLS client configuration.
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.address))
		config, err := network.GetClientTLSConfig(certPath)
		if err != nil {
			log.Fatalf("failed to load certificate (client): %v", err)
		}

		// Execute WithdrawClient.
		client := new(network.WithdrawalClient).New(flags.address, store, config)
		if err := client.Execute(); err != nil {
			log.Fatal(err)
		}
	},
}

// wgUser.
var wgUser sync.WaitGroup

// user charge
var charge = &cobra.Command{
	Use:   "charge  --user USER --bank BANKNAME",
	Short: "USER starts payment server.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}

		// Bind to a bank account.
		if len(flags.bank) == 0 {
			return fmt.Errorf("required \"bank\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}
		store.BankName = flags.bank

		// Load TLS server configuration.
		keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", flags.user))
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.user))
		config, err := network.GetServerTLSConfig(certPath, keyPath)
		if err != nil {
			log.Fatalf("failed to load certificate (server): %v", err)
		}

		// Start GetServer.
		getServer := new(network.GetServer).New(certPath)
		wgUser.Add(1)
		go func() {
			defer wgUser.Done()
			if err := getServer.Start(); err != nil {
				log.Fatalf("failed to start GetServer: %v", err)
			}
		}()

		// Start PaymentServer.
		wgUser.Add(1)
		paymentServer := new(network.PaymentServer).New(store, config)
		go func() {
			defer wgUser.Done()
			if err := paymentServer.Start(); err != nil {
				log.Fatalf("failed to start PaymentServer: %v", err)
			}
		}()

		// Don't exit main thread.
		wgUser.Wait()
	},
}

// user pay
var pay = &cobra.Command{
	Use:   "pay --user USER --server SERVER --bank BANKNAME",
	Short: "USER sends 1 coin to another user at SERVER.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}

		if len(flags.address) == 0 {
			return fmt.Errorf("required \"server\" flag not set")
		}

		if len(flags.bank) == 0 {
			return fmt.Errorf("required \"bank\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}
		store.BankName = flags.bank

		// Execute GetClient.
		setupClient := new(network.GetClient).New(flags.address)
		if err := setupClient.Execute(); err != nil {
			log.Fatal(err)
		}

		// Load TLS client configuration.
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.address))
		config, err := network.GetClientTLSConfig(certPath)
		if err != nil {
			log.Fatalf("failed to load certificate (client): %v", err)
		}

		// Execute PaymentClient.
		paymentClient := new(network.PaymentClient).New(flags.address, store, config)
		if err := paymentClient.Execute(); err != nil {
			log.Fatal(err)
		}
	},
}

// user deposit
var deposit = &cobra.Command{
	Use:   "deposit --user USER --server SERVER",
	Short: "Deposit 1 coin to USER's client account at SERVER.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}

		if len(flags.address) == 0 {
			return fmt.Errorf("required \"server\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		// Execute SetupClient.
		setupClient := new(network.SetupClient).New(flags.address, store)
		if err := setupClient.Execute(); err != nil {
			log.Fatal(err)
		}

		// Load TLS client configuration.
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.address))
		config, err := network.GetClientTLSConfig(certPath)
		if err != nil {
			log.Fatalf("failed to load certificate (client): %v", err)
		}

		// Execute DepositClient.
		depositClient := new(network.DepositClient).New(flags.address, store, config)
		if err := depositClient.Execute(); err != nil {
			log.Fatal(err)
		}
	},
}

// user exchange
var exchange = &cobra.Command{
	Use:   "exchange --user USER --server SERVER",
	Short: "Exchanges an old coin for a new one.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}

		if len(flags.address) == 0 {
			return fmt.Errorf("required \"server\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		// Execute SetupClient.
		setupClient := new(network.SetupClient).New(flags.address, store)
		if err := setupClient.Execute(); err != nil {
			log.Fatal(err)
		}

		// Load TLS client configuration.
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.address))
		config, err := network.GetClientTLSConfig(certPath)
		if err != nil {
			log.Fatalf("failed to load certificate (client): %v", err)
		}

		// Execute ExchangeClient.
		exchangeClient := new(network.ExchangeClient).New(flags.address, store, config)
		if err := exchangeClient.Execute(); err != nil {
			log.Fatal(err)
		}
	},
}

// user inspect
var userInspect = &cobra.Command{
	Use:   "inspect [-f]",
	Short: "View database information.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.user) == 0 {
			return fmt.Errorf("required \"user\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given user: %s", flags.user)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.user))
		store, err := new(store.ClientStore).New(dbPath)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		// Inspect.
		if flags.inspect {
			store.InspectFull()
		} else {
			store.Inspect()
		}
	},
}

// bank
var bank = &cobra.Command{
	Use:   "bank operation",
	Short: "Perform bank operations.",
}

// bank init
var bankInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize ziba system in current computer (as a bank).",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(flags.bank) == 0 {
			return fmt.Errorf("required \"bank\" flag not set")
		}
		if len(flags.identity) == 0 {
			flags.identity = "main"
			// return fmt.Errorf("required \"identity\" flag not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve Ziba directory: %v", err)
		}

		// Create Bank.
		bank := new(core.Bank).New(core.Params)

		// Create local database.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.bank))
		store, err := new(store.BankStore).New(dbPath, flags.identity)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}

		// Write Bank into database.
		store.WriteBank(bank, flags.bank)

		// Create certificates.
		network.CreateCertificate(directory, flags.bank)
	},
}

// wgBank.
var wgBank sync.WaitGroup

// bank serve
var serve = &cobra.Command{
	Use:   "serve",
	Short: "Start servers.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.bank) == 0 {
			return fmt.Errorf("required \"bank\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.bank))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given name: %s", flags.bank)
			}
		}

		if len(flags.identity) == 0 {
			flags.identity = "main"
			// return fmt.Errorf("required \"identity\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve Ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.bank))
		store, err := new(store.BankStore).New(dbPath, flags.identity)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		log.Printf("Bank's Name is: %s", store.Name)

		// Load TLS server configuration.
		keyPath := filepath.Join(directory, fmt.Sprintf("%s_key.pem", flags.bank))
		certPath := filepath.Join(directory, fmt.Sprintf("%s_cert.pem", flags.bank))
		config, err := network.GetServerTLSConfig(certPath, keyPath)
		if err != nil {
			log.Printf("failed to load certificate and key (server): %v", err)
		}

		// Start SetupServer.
		setupServer := new(network.SetupServer).New(store)
		wgBank.Add(1)
		go func() {
			defer wgBank.Done()
			if err := setupServer.Start(); err != nil {
				log.Fatalf("failed to start SetupServer: %v", err)
			}
		}()

		// Start AccgenServer.
		accgenServer := new(network.AccgenServer).New(store, config)
		wgBank.Add(1)
		go func() {
			defer wgBank.Done()
			if err := accgenServer.Start(); err != nil {
				log.Fatalf("failed to start AccgenServer: %v", err)
			}
		}()

		// Start WithdrawalServer.
		withdrawalServer := new(network.WithdrawalServer).New(store, config)
		wgBank.Add(1)
		go func() {
			defer wgBank.Done()
			if err := withdrawalServer.Start(); err != nil {
				log.Fatalf("failed to start WithdrawalServer: %v", err)
			}
		}()

		// Start DepositServer.
		depositServer := new(network.DepositServer).New(store, config)
		wgBank.Add(1)
		go func() {
			defer wgBank.Done()
			if err := depositServer.Start(); err != nil {
				log.Fatalf("failed to start DepositServer: %v", err)
			}
		}()

		// Start ExchangeServer.
		exchangeServer := new(network.ExchangeServer).New(store, config)
		wgBank.Add(1)
		go func() {
			defer wgBank.Done()
			if err := exchangeServer.Start(); err != nil {
				log.Fatalf("failed to start ExchangeServer: %v", err)
			}
		}()

		// Don't exit main thread.
		wgBank.Wait()
	},
}

// bank inspect
var bankInspect = &cobra.Command{
	Use:   "inspect",
	Short: "View database information.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that database file exists.
		if len(flags.bank) == 0 {
			return fmt.Errorf("required \"bank\" flag not set")
		} else {
			directory, err := store.GetZibaDir()
			if err != nil {
				return err
			}
			dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.bank))
			_, err = os.Stat(dbPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("a database file does not exists for given name: %s", flags.bank)
			}
		}

		if len(flags.identity) == 0 {
			flags.identity = "main"
			// return fmt.Errorf("required \"identity\" flag not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get ziba directory.
		directory, err := store.GetZibaDir()
		if err != nil {
			log.Fatalf("failed to retrieve Ziba directory: %v", err)
		}

		// Create store.
		dbPath := filepath.Join(directory, fmt.Sprintf("%s.db", flags.bank))
		store, err := new(store.BankStore).New(dbPath, flags.identity)
		if err != nil {
			log.Fatalf("failed to create store: %v", err)
		}

		// Inspect.
		if flags.inspect {
			store.InspectFull()
		} else {
			store.Inspect()
		}
	},
}

func init() {
	// Global.
	cobra.EnableCommandSorting = false

	// ziba
	ziba.PersistentFlags().StringVarP(&flags.address, "server", "s", "", "Remote server address.")
	ziba.PersistentFlags().StringVarP(&flags.bank, "bank", "b", "", "Bank's name.")
	ziba.PersistentFlags().StringVarP(&flags.user, "user", "u", "", "User's name.")

	// ziba user
	ziba.AddCommand(user)
	// ziba user init
	user.AddCommand(userInit)
	// ziba user accgen
	user.AddCommand(accgen)
	// ziba user withdraw
	user.AddCommand(withdraw)
	// ziba user charge
	user.AddCommand(charge)
	// ziba user pay
	user.AddCommand(pay)
	// ziba user deposit
	user.AddCommand(deposit)
	// ziba user exchange
	user.AddCommand(exchange)
	// ziba user inspect
	user.AddCommand(userInspect)
	userInspect.Flags().BoolVarP(&flags.inspect, "full", "f", false, "Show all fields.")

	// ziba bank
	ziba.AddCommand(bank)
	// ziba bank init
	bank.AddCommand(bankInit)
	// ziba bank serve
	bank.AddCommand(serve)
	// ziba bank inspect
	bank.AddCommand(bankInspect)
	bankInspect.Flags().BoolVarP(&flags.inspect, "full", "f", false, "Show all fields.")
}

func Execute() {
	ziba.Execute()
}
