package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"ziba/core"

	_ "modernc.org/sqlite"
)

// New allocates and returns a new Bankstore for a certain identity.
func (store *BankStore) New(dbPath, identity string) (*BankStore, error) {
	// Get database connection.
	db, err := openDatabase(dbPath)
	if err != nil {
		log.Printf("failed to open database: %v", err)
		return nil, err
	}

	// Grab name.
	var name string
	db.QueryRow(`SELECT name FROM Bank WHERE identity = ?`, identity).Scan(&name)

	// Keep values.
	store.db = db
	store.Name = name
	store.identity = identity

	// Init schema.
	err = store.createTables()
	if err != nil {
		log.Fatalf("failed to create Bank's database schema: %v", err)
		return nil, err
	}

	// Create store.
	return store, nil
}

// CreateTables creates the database schema for a bank's local database.
// Only creates the tables if they don't previously exist.
func (store *BankStore) createTables() error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	table := `CREATE TABLE IF NOT EXISTS Bank (
	-- keys
	id 	 		 INTEGER PRIMARY KEY AUTOINCREMENT,
	name		 TEXT NOT NULL,
	identity TEXT UNIQUE ON CONFLICT IGNORE NOT NULL,

	-- Bank
	Priv TEXT NOT NULL,
	Pub  TEXT NOT NULL,
	---- SchemeParams
	scheme_Q TEXT NOT NULL,
	scheme_P TEXT NOT NULL,
	scheme_G TEXT NOT NULL,
	---- RsaKey
	key_P TEXT NOT NULL,
	key_Q TEXT NOT NULL,
	key_D TEXT NOT NULL,
	key_N TEXT NOT NULL,
	key_E TEXT NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS ClientInfo (
	-- keys
	id 	 INTEGER PRIMARY KEY AUTOINCREMENT,
	hash INTEGER UNIQUE ON CONFLICT IGNORE NOT NULL, -- ClientProfile hash

	-- ClientInfo
	K 				 TEXT NOT NULL,
	S 				 TEXT NOT NULL,
	Credential TEXT NOT NULL,
	Contract 	 TEXT NOT NULL,
	---- ClientProfile
	PrivStamp 	 TEXT NOT NULL,
	IdentityHash TEXT NOT NULL,
	TradeId			 TEXT NOT NULL,
	Pub 				 TEXT NOT NULL,
	N 					 TEXT NOT NULL,
	E 					 TEXT NOT NULL, 
	
	balance INTEGER NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS CoinProfile (
	-- keys
	id 	 INTEGER PRIMARY KEY AUTOINCREMENT,
	hash INTEGER UNIQUE ON CONFLICT IGNORE NOT NULL, -- CoinProfile hash

	-- CoinProfile
	Pub 			 TEXT NOT NULL,
	First 		 TEXT NOT NULL,
	A  				 TEXT NOT NULL,
	R  				 TEXT NOT NULL,
	A2 				 TEXT NOT NULL,
	Expiration DATETIME NOT NULL,
	Second 		 TEXT NOT NULL,
	Msg 			 TEXT NOT NULL,

	operation INTEGER NOT NULL,
	client 	 	INTEGER NOT NULL, -- ClientProfile hash
	date 	 	 	DATETIME NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// WriteBank attempts to write bank into the local database.
// If an entry exists for this BankStore's identity nothing is written into the database.
func (store *BankStore) WriteBank(bank *core.Bank, name string) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Associate Bank's name.
	store.Name = name

	// Check if an identity already exists.
	var id int64
	err = tx.QueryRow(`SELECT id FROM Bank WHERE identity = ?`, store.identity).Scan(&id)
	if err != sql.ErrNoRows {
		log.Printf("a bank (id: %d) already exists for identity %s", id, store.identity)
		return nil
	}

	stmt := `INSERT INTO
	Bank 	 (identity, name, Priv, Pub, scheme_Q, scheme_P, scheme_G, key_P, key_Q, key_D, key_N, key_E)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		store.identity,
		store.Name,
		toString(bank.Priv),
		toString(bank.Pub),
		toString(bank.Scheme.Q),
		toString(bank.Scheme.P),
		toString(bank.Scheme.G),
		toString(bank.Key.P),
		toString(bank.Key.Q),
		toString(bank.Key.D),
		toString(bank.Key.N),
		toString(bank.Key.E),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ReadBank attempts to read the entry for this BankStore's identity.
// If no entry exists the return value is nil.
func (store *BankStore) ReadBank() (*core.Bank, error) {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	stmt := `SELECT Priv, Pub, scheme_Q, scheme_P, scheme_G, key_P, key_Q, key_D, key_N, key_E FROM Bank WHERE identity = ?`
	scanner := new(rowScanner).New(10)
	err = tx.QueryRow(stmt, store.identity).Scan(scanner.dest...)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	} else if err != nil {
		return nil, err
	}
	vals := scanner.Strings()
	bank := &core.Bank{
		Priv: fromString(vals[0]),
		Pub:  fromString(vals[1]),
		Scheme: core.SchemeParams{
			Q: fromString(vals[2]),
			P: fromString(vals[3]),
			G: fromString(vals[4]),
		},
		Key: core.RsaKey{
			P: fromString(vals[5]),
			Q: fromString(vals[6]),
			D: fromString(vals[7]),
			N: fromString(vals[8]),
			E: fromString(vals[9]),
		},
	}

	return bank, tx.Commit()
}

// WriteClientInfo attempts to write client into the local database.
// If an entry exists for the client's profile hash, ErrExistingClient is returned.
func (store *BankStore) WriteClientInfo(client *core.ClientInfo) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Check if this client already exists.
	var id int64
	err = tx.QueryRow(`SELECT id FROM ClientInfo WHERE hash = ?`, client.Profile.Hash()).Scan(&id)
	if err != sql.ErrNoRows {
		log.Printf("a client (id: %d) already exists", id)
		return ErrExistingClient
	}

	stmt := `INSERT INTO
	ClientInfo (hash, K, S, Credential, Contract, PrivStamp, IdentityHash, TradeId, Pub, N, E, balance)
	VALUES 		 (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		client.Profile.Hash(),
		toString(client.K),
		toString(client.S),
		toString(client.Credential),
		toString(client.Contract),
		toString(client.Profile.PrivStamp),
		toString(client.Profile.IdentityHash),
		toString(client.Profile.TradeId),
		toString(client.Profile.Pub),
		toString(client.Profile.N),
		toString(client.Profile.E),
		100,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ReadClientInfo attempts to read the entry for this client's profile hash.
// Returns sql.ErrNoRows if no entry exists.
func (store *BankStore) ReadClientInfo(client *core.ClientProfile) (*core.ClientInfo, error) {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// Check if this client already exists.
	var id int64
	err = tx.QueryRow(`SELECT id FROM ClientInfo WHERE hash = ?`, client.Hash()).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	} else if err != nil {
		return nil, err
	}

	stmt := `SELECT K, S, Credential, Contract FROM ClientInfo WHERE hash = ?`
	scanner := new(rowScanner).New(4)
	err = tx.QueryRow(stmt, client.Hash()).Scan(scanner.dest...)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	vals := scanner.Strings()
	clientInfo := &core.ClientInfo{
		Profile:    *client,
		K:          fromString(vals[0]),
		S:          fromString(vals[1]),
		Credential: fromString(vals[2]),
		Contract:   fromString(vals[3]),
	}

	return clientInfo, tx.Commit()
}

// ReadClientBalance.
func (store *BankStore) ReadClientBalance(client *core.ClientProfile) (int64, error) {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return 0, err
	}
	defer tx.Rollback()

	var balance int64
	stmt := `SELECT balance FROM ClientInfo WHERE hash = ?`
	err = tx.QueryRow(stmt, client.Hash()).Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, tx.Commit()
}

// UpdateClientBalance.
func (store *BankStore) UpdateClientBalance(client *core.ClientProfile, balance int64) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	stmt := `UPDATE ClientInfo SET balance = ? WHERE hash = ?`
	_, err = tx.Exec(stmt, balance, client.Hash())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// WriteCoinProfile attempts to write coin into the local database.
// If an entry exists for the coin's profile hash, ErrExistingCoin is returned.
func (store *BankStore) WriteCoinProfile(coin *core.CoinProfile, operation Operation_Type, client *core.ClientProfile) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Check if this coin already exists.
	var id int64
	err = tx.QueryRow(`SELECT id FROM CoinProfile WHERE hash = ?`, coin.Hash()).Scan(&id)
	if err != sql.ErrNoRows {
		log.Printf("a coin (id: %d) already exists", id)
		return ErrExistingCoin
	}

	stmt := `INSERT INTO
	CoinProfile (hash, Pub, First, A, R, A2, Expiration, Second, Msg, operation, client, date)
	VALUES			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		coin.Hash(),
		toString(coin.Pub),
		toString(coin.First),
		toString(coin.A),
		toString(coin.R),
		toString(coin.A2),
		coin.Expiration,
		toString(coin.Second),
		toString(coin.Msg),
		operation,
		client.Hash(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ReadCoinProfile attempts to read the entry for this coin's profile hash.
// Returns sql.ErrNoRows if no entry exists.
func (store *BankStore) ReadCoinProfile(coin *core.CoinProfile) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Check if this coin already exists.
	var id int64
	err = tx.QueryRow(`SELECT id FROM CoinProfile WHERE hash = ?`, coin.Hash()).Scan(&id)
	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	} else {
		return err
	}
}

// Inspect.
func (store *BankStore) Inspect() {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Fatalf("failed to initiate transaction: %v", err)
	}
	defer tx.Rollback()

	// Bank.
	fmt.Printf("\nBANK\n")
	rows, err := tx.Query(`SELECT id, name, identity FROM Bank`)
	if err != nil {
		log.Fatalf("failed to query Bank table: %v", err)
	}
	fmt.Printf("%-5s %-10s %-10s\n", "ID", "Name", "Identity")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			name     string
			identity string
		)

		err = rows.Scan(&id, &name, &identity)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		fmt.Printf("%-5d %-10s %-10s\n", id, name, identity)
	}

	// ClientInfo.
	fmt.Printf("\nCLIENT INFO\n")
	rows, err = tx.Query(`SELECT id, hash, balance FROM ClientInfo`)
	if err != nil {
		log.Fatalf("failed to query ClientInfo table: %v", err)
	}
	fmt.Printf("%-5s %-10s %-10s\n", "ID", "ClientHash", "Balance")
	for rows.Next() {
		// Scanner variables.
		var (
			id      int64
			client  int64
			balance int64
		)

		err = rows.Scan(&id, &client, &balance)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		fmt.Printf("%-5d %-10d %-10d\n", id, client, balance)
	}

	// CoinProfile.
	fmt.Printf("\nCOIN PROFILE\n")
	rows, err = tx.Query(`SELECT id, hash, operation, client, date FROM CoinProfile`)
	if err != nil {
		log.Fatalf("failed to query CoinProfile table: %v", err)
	}
	fmt.Printf("%-5s %-10s %-10s %-10s %-23s\n", "ID", "CoinHash", "Operation", "ClientHash", "Date")
	for rows.Next() {
		// Scanner variables.
		var (
			id         int64
			coinHash   int64
			operation  Operation_Type
			clientHash int64
			date       time.Time
		)

		err = rows.Scan(&id, &coinHash, &operation, &clientHash, &date)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		var operationStr string
		switch operation {
		case Operation_Deposit:
			operationStr = "Deposit"
		case Operation_Exchange:
			operationStr = "Exchange"
		default:
		}

		fmt.Printf("%-5d %-10.10d %-10s %-10.10d %-23s\n", id, coinHash, operationStr, clientHash, date.String()[:23])
	}

	// Commit transaction.
	if err := tx.Commit(); err != nil {
		log.Fatalf("failed to commit transaction: %v", err)
	}
}

// InspectFull.
func (store *BankStore) InspectFull() {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Fatalf("failed to initiate transaction: %v", err)
	}
	defer tx.Rollback()

	// Bank.
	fmt.Printf("\nBANK\n")
	rows, err := tx.Query(`SELECT id, name, identity, Priv, Pub, scheme_Q, scheme_P, scheme_G, key_P, key_Q, key_D, key_N, key_E FROM Bank`)
	if err != nil {
		log.Fatalf("failed to query Bank table: %v", err)
	}
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "Name", "Identity", "Priv", "Pub", "Scheme:Q", "Scheme:P", "Scheme:G", "Key:P", "Key:Q", "Key:D", "Key:N", "Key:E")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			name     string
			identity string
			numbers  [2]string
			scheme   [3]string
			key      [5]string
		)

		err = rows.Scan(&id, &name, &identity, &numbers[0], &numbers[1], &scheme[0], &scheme[1], &scheme[2], &key[0], &key[1], &key[2], &key[3], &key[4])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		fmt.Printf("%-5d %-10s %-10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, name, identity, numbers[0], numbers[1], scheme[0], scheme[1], scheme[2], key[0], key[1], key[2], key[3], key[4])
	}

	// ClientInfo.
	fmt.Printf("\nCLIENT INFO\n")
	rows, err = tx.Query(`SELECT id, hash, balance, K, S, Credential, Contract, PrivStamp, IdentityHash, TradeId, Pub, N, E FROM ClientInfo`)
	if err != nil {
		log.Fatalf("failed to query ClientInfo table: %v", err)
	}
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "ClientHash", "Balance", "K", "S", "Credential", "Contract", "PrivStamp", "IdHash", "TradeId", "Pub", "N", "E")
	for rows.Next() {
		// Scanner variables.
		var (
			id         int64
			clientHash int64
			balance    int64
			info       [4]string
			profile    [6]string
		)

		err = rows.Scan(&id, &clientHash, &balance, &info[0], &info[1], &info[2], &info[3], &profile[0], &profile[1], &profile[2], &profile[3], &profile[4], &profile[5])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		fmt.Printf("%-5d %-10d %-10d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, clientHash, balance, info[0], info[1], info[2], info[3], profile[0], profile[1], profile[2], profile[3], profile[4], profile[5])
	}

	// CoinProfile.
	fmt.Printf("\nCOIN PROFILE\n")
	rows, err = tx.Query(`SELECT id, hash, Pub, First, A, R, A2, Expiration, Second, Msg, operation, client, date FROM CoinProfile`)
	if err != nil {
		log.Fatalf("failed to query CoinProfile table: %v", err)
	}
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s %-23s %-11s %-10s %-10s %-10s %-23s\n", "ID", "CoinHash", "Coin:Pub", "Coin:First", "Coin:A", "Coin:R", "Coin:A2", "Coin:Expiration", "Coin:Second", "Coin:Msg", "Operation", "ClientHash", "Date")
	for rows.Next() {
		// Scanner variables.
		var (
			id         int64
			coinHash   int64
			profile    [7]string
			expiration time.Time
			operation  Operation_Type
			clientHash int64
			date       time.Time
		)

		err = rows.Scan(&id, &coinHash, &profile[0], &profile[1], &profile[2], &profile[3], &profile[4], &expiration, &profile[5], &profile[6], &operation, &clientHash, &date)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		var operationStr string
		switch operation {
		case Operation_Deposit:
			operationStr = "Deposit"
		case Operation_Exchange:
			operationStr = "Exchange"
		default:
		}

		fmt.Printf("%-5d %-10.10d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-23s %-11.11s %-10.10s %-10s %-10.10d %-23s\n", id, coinHash, profile[0], profile[1], profile[2], profile[3], profile[4], expiration.String()[:23], profile[5], profile[6], operationStr, clientHash, date.String()[:23])
	}

	// Commit transaction.
	if err := tx.Commit(); err != nil {
		log.Fatalf("failed to commit transaction: %v", err)
	}
}
