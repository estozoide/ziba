package store

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
	"ziba/core"

	_ "modernc.org/sqlite"
)

// New allocates and returns a new ClientStore for a bank identified by bankName.
func (store *ClientStore) New(dbPath string) (*ClientStore, error) {
	// Get database connection.
	db, err := openDatabase(dbPath)
	if err != nil {
		log.Printf("failed to open database: %v", err)
		return nil, err
	}
	store.db = db

	// Init tables.
	err = store.createTables()
	if err != nil {
		log.Fatalf("failed to create User's database schema: %v", err)
		return nil, err
	}

	// Create store.
	return store, nil
}

// CreateTables creates the database schema for a bank's local database.
// Only creates the tables if they don't previously exist.
func (store *ClientStore) createTables() error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	table := `CREATE TABLE IF NOT EXISTS Client (
	-- keys
	id 	 INTEGER PRIMARY KEY AUTOINCREMENT,
	bank TEXT UNIQUE ON CONFLICT IGNORE NOT NULL,

	-- Client
	TradeId 	 TEXT NOT NULL,
	Priv		 	 TEXT NOT NULL,
	Pub 			 TEXT NOT NULL,
	Credential TEXT NOT NULL,
	Contract 	 TEXT NOT NULL,
	---- BankProfile
	---- RsaKey

	localBalance  INTEGER NOT NULL,
	remoteBalance INTEGER NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS BankProfile (
	-- keys
	id 		 INTEGER PRIMARY KEY AUTOINCREMENT,
	client INTEGER UNIQUE ON CONFLICT IGNORE REFERENCES Client(id) ON DELETE CASCADE,

	-- BankProfile
	Pub TEXT NOT NULL,
	N 	TEXT NOT NULL,
	E 	TEXT NOT NULL,
	---- SchemeParams
	Q TEXT NOT NULL,
	P TEXT NOT NULL,
	G TEXT NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS RsaKey (
	-- keys
	id 		 INTEGER PRIMARY KEY AUTOINCREMENT,
	client INTEGER UNIQUE ON CONFLICT IGNORE REFERENCES Client(id) ON DELETE CASCADE,

	-- RsaKey
	P TEXT NOT NULL,
	Q TEXT NOT NULL,
	D TEXT NOT NULL,
	N TEXT NOT NULL,
	E TEXT NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS Coin (
	-- keys
	id 		 INTEGER PRIMARY KEY AUTOINCREMENT,
	client INTEGER REFERENCES Client(id) ON DELETE CASCADE,
	hash 	 INTEGER UNIQUE ON CONFLICT IGNORE NOT NULL -- CoinProfile hash
	
	-- Coin
	---- CoinRandom
	---- CoinElgamal
	---- CoinParams
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS CoinRandom (
	-- keys
	id 	 INTEGER PRIMARY KEY AUTOINCREMENT,
	coin INTEGER UNIQUE ON CONFLICT IGNORE REFERENCES Coin(id) ON DELETE CASCADE,

	-- CoinRandom
	E 			 TEXT NOT NULL,
	L 			 TEXT NOT NULL,
	LInv   	 TEXT NOT NULL,
	Beta1 	 TEXT NOT NULL,
	Beta1Inv TEXT NOT NULL,
	Beta2 	 TEXT NOT NULL,
	Y 			 TEXT NOT NULL,
	YInv 		 TEXT NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS CoinElgamal (
	-- keys
	id 	 INTEGER PRIMARY KEY AUTOINCREMENT,
	coin INTEGER UNIQUE ON CONFLICT IGNORE REFERENCES Coin(id) ON DELETE CASCADE,

	-- CoinElgamal
	Priv 	 TEXT NOT NULL,
	Pub 	 TEXT NOT NULL,
	First  TEXT NOT NULL,
	Second TEXT NOT NULL,
	Msg 	 TEXT NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	table = `CREATE TABLE IF NOT EXISTS CoinParams (
	-- keys
	id 	 INTEGER PRIMARY KEY AUTOINCREMENT,
	coin INTEGER UNIQUE ON CONFLICT IGNORE REFERENCES Coin(id) ON DELETE CASCADE,

	-- CoinParams
	A 				 TEXT NOT NULL,
	ALower 		 TEXT NOT NULL,
	C 				 TEXT NOT NULL,
	Expiration DATETIME NOT NULL,
	A1		 		 TEXT NOT NULL,
	C1 				 TEXT NOT NULL,
	A2 				 TEXT NOT NULL,
	R 				 TEXT NOT NULL
	);`
	_, err = tx.Exec(table)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// WriteClient attempts to write client into the local database.
// If an entry exists for this ClientStore's bank nothing is written into the database.
func (store *ClientStore) WriteClient(client *core.Client) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Check if a client already exists for that bank.
	var id int64
	err = tx.QueryRow(`SELECT id FROM Client WHERE bank = ?`, store.BankName).Scan(&id)
	if err != sql.ErrNoRows {
		log.Printf("a client (id: %d) already exists for bank %s", id, store.BankName)
		return nil
	}

	stmt := `INSERT INTO
	Client (bank, TradeId, Priv, Pub, Credential, Contract, localBalance, remoteBalance)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	res, err := tx.Exec(stmt,
		store.BankName,
		toString(client.TradeId),
		toString(client.Priv),
		toString(client.Pub),
		toString(client.Credential),
		toString(client.Contract),
		0,
		100,
	)
	if err != nil {
		return err
	}
	clientId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	stmt = `INSERT INTO
	BankProfile (client, Pub, N, E, Q, P, G)
	VALUES 			(?, ?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		clientId,
		toString(client.Bank.Pub),
		toString(client.Bank.N),
		toString(client.Bank.E),
		toString(client.Bank.Scheme.Q),
		toString(client.Bank.Scheme.P),
		toString(client.Bank.Scheme.G),
	)
	if err != nil {
		return err
	}

	stmt = `INSERT INTO
	RsaKey (client, P, Q, N, D, E)
	VALUES (?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		clientId,
		toString(client.Key.P),
		toString(client.Key.Q),
		toString(client.Key.N),
		toString(client.Key.D),
		toString(client.Key.E),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ReadClient attempts to read the entry for this ClientStore's bank.
// If no entry exists the return value is nil.
func (store *ClientStore) ReadClient() (*core.Client, error) {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	stmt := `SELECT id, TradeId, Priv, Pub, Credential, Contract, localBalance, remoteBalance FROM Client WHERE bank = ?`
	scanner := new(rowScanner).New(8)
	err = tx.QueryRow(stmt, store.BankName).Scan(scanner.dest...)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	vals := scanner.Strings()
	client := &core.Client{
		TradeId:    fromString(vals[1]),
		Priv:       fromString(vals[2]),
		Pub:        fromString(vals[3]),
		Credential: fromString(vals[4]),
		Contract:   fromString(vals[5]),
	}
	// Keep this client's id & balance.
	store.clientId, _ = strconv.ParseInt(vals[0], 10, 64)
	store.LocalBalance, _ = strconv.ParseInt(vals[6], 10, 64)
	store.RemoteBalance, _ = strconv.ParseInt(vals[7], 10, 64)

	stmt = `SELECT P, Q, N, D, E FROM RsaKey WHERE client = ?`
	scanner = new(rowScanner).New(5)
	err = tx.QueryRow(stmt, store.clientId).Scan(scanner.dest...)
	if err != nil {
		return nil, err
	}
	vals = scanner.Strings()
	key := core.RsaKey{
		P: fromString(vals[0]),
		Q: fromString(vals[1]),
		N: fromString(vals[2]),
		D: fromString(vals[3]),
		E: fromString(vals[4]),
	}

	stmt = `SELECT Pub, N, E, Q, P, G FROM BankProfile WHERE client = ?`
	scanner = new(rowScanner).New(6)
	err = tx.QueryRow(stmt, store.clientId).Scan(scanner.dest...)
	if err != nil {
		return nil, err
	}
	vals = scanner.Strings()
	bank := core.BankProfile{
		Scheme: core.SchemeParams{
			Q: fromString(vals[3]),
			P: fromString(vals[4]),
			G: fromString(vals[5]),
		},
		Pub: fromString(vals[0]),
		N:   fromString(vals[1]),
		E:   fromString(vals[2]),
	}

	client.Key = key
	client.Bank = bank

	return client, tx.Commit()
}

// WriteCoin writes coin into the local database.
// Only to be called after a ReadClient call to initialize the client's id of this ClientStore.
func (store *ClientStore) WriteCoin(coin *core.Coin, operation Operation_Type) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	stmt := `INSERT INTO
	Coin 	 (client, hash)
	VALUES (?, ?);`
	res, err := tx.Exec(stmt, store.clientId, coin.Profile().Hash())
	if err != nil {
		return err
	}
	coinId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	stmt = `INSERT INTO
	CoinRandom (coin, E, L, LInv, Beta1, Beta1Inv, Beta2, Y, YInv)
	VALUES		 (?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		coinId,
		toString(coin.Random.E),
		toString(coin.Random.L),
		toString(coin.Random.LInv),
		toString(coin.Random.Beta1),
		toString(coin.Random.Beta1Inv),
		toString(coin.Random.Beta2),
		toString(coin.Random.Y),
		toString(coin.Random.YInv),
	)
	if err != nil {
		return err
	}

	stmt = `INSERT INTO
	CoinElgamal (coin, Priv, Pub, First, Second, Msg)
	VALUES 			(?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		coinId,
		toString(coin.Elgamal.Priv),
		toString(coin.Elgamal.Pub),
		toString(coin.Elgamal.First),
		toString(coin.Elgamal.Second),
		toString(coin.Elgamal.Msg),
	)
	if err != nil {
		return err
	}

	stmt = `INSERT INTO
	CoinParams (coin, A, ALower, C, Expiration, A1, C1, A2, R)
	VALUES 		 (?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err = tx.Exec(stmt,
		coinId,
		toString(coin.Params.A),
		toString(coin.Params.ALower),
		toString(coin.Params.C),
		coin.Params.Expiration,
		toString(coin.Params.A1),
		toString(coin.Params.C1),
		toString(coin.Params.A2),
		toString(coin.Params.R),
	)
	if err != nil {
		return err
	}

	stmt = `UPDATE Client SET localBalance = localBalance + ? WHERE id = ?;`
	_, err = tx.Exec(stmt, 1, store.clientId)
	if err != nil {
		return err
	}

	// Update remote balance given the type of operation.
	switch operation {
	case Operation_Withdrawal:
		stmt = `UPDATE Client Set remoteBalance = remoteBalance - ? WHERE id = ?`
		_, err = tx.Exec(stmt, 1, store.clientId)
		if err != nil {
			return err
		}
	case Operation_Payment:
	case Operation_Deposit:
	case Operation_Exchange:
	default:
	}

	return tx.Commit()
}

// ReadCoins returns a tuple-like struct: a coin object paired with its database coin id.
// Only to be called after a ReadClient call to initialize the client's id of this ClientStore.
func (store *ClientStore) ReadCoins() ([]core.Coin, error) {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	stmt := `SELECT id FROM Coin WHERE client = ?`
	rows, err := tx.Query(stmt, store.clientId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coins []core.Coin

	for rows.Next() {
		var coinId int64

		err := rows.Scan(&coinId)
		if err == sql.ErrNoRows {
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		stmt = `SELECT E, L, LInv, Beta1, Beta1Inv, Beta2, Y, YInv FROM CoinRandom WHERE coin = ?`
		scanner := new(rowScanner).New(8)
		err = tx.QueryRow(stmt, coinId).Scan(scanner.dest...)
		if err != nil {
			return nil, err
		}
		vals := scanner.Strings()
		random := core.CoinRandom{
			E:        fromString(vals[0]),
			L:        fromString(vals[1]),
			LInv:     fromString(vals[2]),
			Beta1:    fromString(vals[3]),
			Beta1Inv: fromString(vals[4]),
			Beta2:    fromString(vals[5]),
			Y:        fromString(vals[6]),
			YInv:     fromString(vals[7]),
		}

		stmt = `SELECT Priv, Pub, First, Second, Msg FROM CoinElgamal WHERE coin = ?`
		scanner = new(rowScanner).New(5)
		err = tx.QueryRow(stmt, coinId).Scan(scanner.dest...)
		if err != nil {
			return nil, err
		}
		vals = scanner.Strings()
		elgamal := core.CoinElgamal{
			Priv:   fromString(vals[0]),
			Pub:    fromString(vals[1]),
			First:  fromString(vals[2]),
			Second: fromString(vals[3]),
			Msg:    fromString(vals[4]),
		}

		stmt = `SELECT A, ALower, C, Expiration, A1, C1, A2, R FROM CoinParams WHERE coin = ?`
		scanner = new(rowScanner).New(8)
		err = tx.QueryRow(stmt, coinId).Scan(scanner.dest...)
		if err != nil {
			return nil, err
		}
		vals = scanner.Strings()
		expiration, _ := time.Parse(time.RFC3339, vals[3])
		params := core.CoinParams{
			A:          fromString(vals[0]),
			ALower:     fromString(vals[1]),
			C:          fromString(vals[2]),
			Expiration: expiration,
			A1:         fromString(vals[4]),
			C1:         fromString(vals[5]),
			A2:         fromString(vals[6]),
			R:          fromString(vals[7]),
		}

		coin := core.Coin{
			Random:  random,
			Elgamal: elgamal,
			Params:  params,
		}

		coins = append(coins, coin)
	}

	return coins, tx.Commit()
}

// DeleteCoin deletes a coin entry (and its dependencies) given a coin id retrieved by a ReadCoins call.
func (store *ClientStore) DeleteCoin(coin *core.Coin, operation Operation_Type) error {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Printf("failed to initiate transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	stmt := `DELETE FROM Coin WHERE hash = ?`
	_, err = tx.Exec(stmt, coin.Profile().Hash())
	if err != nil {
		return err
	}

	stmt = `UPDATE Client SET localBalance = localBalance - ? WHERE id = ?;`
	_, err = tx.Exec(stmt, 1, store.clientId)
	if err != nil {
		return err
	}

	// Update remote balance given the type of operation.
	switch operation {
	case Operation_Withdrawal:
	case Operation_Payment:
	case Operation_Deposit:
		stmt = `UPDATE Client Set remoteBalance = remoteBalance + ? WHERE id = ?`
		_, err = tx.Exec(stmt, 1, store.clientId)
		if err != nil {
			return err
		}
	case Operation_Exchange:
	default:
	}

	return tx.Commit()
}

// Inspect.
func (store *ClientStore) Inspect() {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Fatalf("failed to initiate transaction: %v", err)
	}
	defer tx.Rollback()

	// Client.
	fmt.Printf("\nCLIENT\n")
	rows, err := tx.Query(`SELECT id, bank, localBalance, remoteBalance FROM Client`)
	if err != nil {
		log.Fatalf("failed to query Client: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s\n", "ID", "Bank", "Local", "Remote")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			bankName string
			local    int64
			remote   int64
		)

		err = rows.Scan(&id, &bankName, &local, &remote)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10s $%-9d $%-9d\n", id, bankName, local, remote)
	}

	// Coin.
	fmt.Printf("\nCOIN\n")
	rows, err = tx.Query(`SELECT Coin.id, Coin.hash, Client.bank FROM Coin JOIN Client ON Coin.client = Client.id`)
	if err != nil {
		log.Fatalf("failed to query Coin: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s\n", "ID", "CoinHash", "Bank")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			coinHash int64
			bankName string
		)

		err = rows.Scan(&id, &coinHash, &bankName)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10.10d %-10s\n", id, coinHash, bankName)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("failed to commit transaction: %v", err)
	}
}

// InspectFull.
func (store *ClientStore) InspectFull() {
	// Begin a transaction.
	tx, err := store.db.Begin()
	if err != nil {
		log.Fatalf("failed to initiate transaction: %v", err)
	}
	defer tx.Rollback()

	// Client.
	fmt.Printf("\nCLIENT\n")
	rows, err := tx.Query(`SELECT id, bank, localBalance, remoteBalance, TradeId, Priv, Pub, Credential, Contract FROM Client`)
	if err != nil {
		log.Fatalf("failed to query Client: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "Bank", "Local", "Remote", "TradeId", "Priv", "Pub", "Credential", "Contract")
	for rows.Next() {
		// Scanner variables.
		var (
			id            int64
			bankName      string
			localBalance  int64
			remoteBalance int64
			client        [5]string
		)

		err = rows.Scan(&id, &bankName, &localBalance, &remoteBalance, &client[0], &client[1], &client[2], &client[3], &client[4])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10s $%-9d $%-9d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, bankName, localBalance, remoteBalance, client[0], client[1], client[2], client[3], client[4])
	}

	// BankProfile.
	fmt.Printf("\nBANK PROFILE\n")
	rows, err = tx.Query(`SELECT id, client, Pub, N, E, Q, P, G FROM BankProfile`)
	if err != nil {
		log.Fatalf("failed to query BankProfile: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "ClientId", "Bank:Pub", "Bank:N", "Bank:E", "Scheme:Q", "Scheme:P", "Scheme:G")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			clientId int64
			profile  [3]string
			scheme   [3]string
		)

		err = rows.Scan(&id, &clientId, &profile[0], &profile[1], &profile[2], &scheme[0], &scheme[1], &scheme[2])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, clientId, profile[0], profile[1], profile[2], scheme[0], scheme[1], scheme[2])
	}

	// RsaKey.
	fmt.Printf("\nRSA KEY\n")
	rows, err = tx.Query(`SELECT id, client, P, Q, D, N, E FROM RsaKey`)
	if err != nil {
		log.Fatalf("failed to query RsaKey: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "ClientId", "P", "Q", "D", "N", "E")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			clientId int64
			rsaKey   [5]string
		)

		err = rows.Scan(&id, &clientId, &rsaKey[0], &rsaKey[1], &rsaKey[2], &rsaKey[3], &rsaKey[4])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, clientId, rsaKey[0], rsaKey[1], rsaKey[2], rsaKey[3], rsaKey[4])
	}

	// Coin.
	fmt.Printf("\nCOIN\n")
	rows, err = tx.Query(`SELECT id, client, hash FROM Coin`)
	if err != nil {
		log.Fatalf("failed to query Coin: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s\n", "ID", "ClientId", "CoinHash")
	for rows.Next() {
		// Scanner variables.
		var (
			id       int64
			clientId int64
			coinHash int64
		)

		err = rows.Scan(&id, &clientId, &coinHash)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10d %-10.10d\n", id, clientId, coinHash)
	}

	// CoinRandom.
	fmt.Printf("\nCOIN RANDOM\n")
	rows, err = tx.Query(`SELECT id, coin, E, L, LInv, Beta1, Beta1Inv, Beta2, Y, YInv FROM CoinRandom`)
	if err != nil {
		log.Fatalf("failed to query CoinRandom: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "CoinId", "E", "L", "LInv", "Beta1", "Beta1Inv", "Beta2", "Y", "YInv")
	for rows.Next() {
		// Scanner variables.
		var (
			id     int64
			coinId int64
			random [8]string
		)

		err = rows.Scan(&id, &coinId, &random[0], &random[1], &random[2], &random[3], &random[4], &random[5], &random[6], &random[7])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, coinId, random[0], random[1], random[2], random[3], random[4], random[5], random[6], random[7])
	}

	// CoinElgamal.
	fmt.Printf("\nCOIN ELGAMAL\n")
	rows, err = tx.Query(`SELECT id, coin, Priv, Pub, First, Second, Msg FROM CoinElgamal`)
	if err != nil {
		log.Fatalf("failed to query CoinElgamal: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-10s %-10s\n", "ID", "CoinId", "Priv", "Pub", "First", "Second", "Msg")
	for rows.Next() {
		// Scanner variables.
		var (
			id      int64
			coinId  int64
			elgamal [5]string
		)

		err = rows.Scan(&id, &coinId, &elgamal[0], &elgamal[1], &elgamal[2], &elgamal[3], &elgamal[4])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10d %-10.10s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, coinId, elgamal[0], elgamal[1], elgamal[2], elgamal[3], elgamal[4])
	}

	// CoinParams.
	fmt.Printf("\nCOIN PARAMS\n")
	rows, err = tx.Query(`SELECT id, coin, A, ALower, C, Expiration, A1, C1, A2, R FROM CoinParams`)
	if err != nil {
		log.Fatalf("failed to query CoinParams: %v", err)
	}
	// Print output header.
	fmt.Printf("%-5s %-10s %-10s %-10s %-10s %-23s %-10s %-10s %-10s %-10s\n", "ID", "CoinId", "A", "ALower", "C", "Expiration", "A1", "C1", "A2", "R")
	for rows.Next() {
		// Scanner variables.
		var (
			id         int64
			coinId     int64
			params     [7]string
			expiration time.Time
		)

		err = rows.Scan(&id, &coinId, &params[0], &params[1], &params[2], &expiration, &params[3], &params[4], &params[5], &params[6])
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatalf("failed to scan: %v", err)
		}

		// Print output row.
		fmt.Printf("%-5d %-10d %-10.10s %-10.10s %-10.10s %-23.23s %-10.10s %-10.10s %-10.10s %-10.10s\n", id, coinId, params[0], params[1], params[2], expiration.String(), params[3], params[4], params[5], params[6])
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("failed to commit transaction: %v", err)
	}
}
