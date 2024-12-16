package network

import (
	"crypto/tls"
	"ziba/store"
)

//
// SETUP
//

// SetupServer.
type SetupServer struct {
	port  int
	store *store.BankStore
}

// SetupClient.
type SetupClient struct {
	serverAddr string
	store      *store.ClientStore
}

//
// ACCOUNT GENERATION
//

// AccgenServer.
type AccgenServer struct {
	port   int
	store  *store.BankStore
	config *tls.Config
}

// AccgenClient.
type AccgenClient struct {
	serverAddr string
	store      *store.ClientStore
	config     *tls.Config
}

//
// WITHDRAWAL
//

// WithdrawalServer.
type WithdrawalServer struct {
	port   int
	store  *store.BankStore
	config *tls.Config
}

// WithdrawalClient.
type WithdrawalClient struct {
	serverAddr string
	store      *store.ClientStore
	config     *tls.Config
}

// PaymentServer.
type PaymentServer struct {
	port   int
	store  *store.ClientStore
	config *tls.Config
}

// PaymentClient.
type PaymentClient struct {
	serverAddr string
	store      *store.ClientStore
	config     *tls.Config
}

// DepositServer.
type DepositServer struct {
	port   int
	store  *store.BankStore
	config *tls.Config
}

// DepositClient.
type DepositClient struct {
	serverAddr string
	store      *store.ClientStore
	config     *tls.Config
}

// ExchangeServer.
type ExchangeServer struct {
	port   int
	store  *store.BankStore
	config *tls.Config
}

// ExchangeClient.
type ExchangeClient struct {
	serverAddr string
	store      *store.ClientStore
	config     *tls.Config
}

// GetServer.
type GetServer struct {
	port     int
	filepath string
}

// GetClient.
type GetClient struct {
	serverAddr string
}
