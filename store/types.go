package store

import (
	"database/sql"
)

// ClientStore handles a client's local database operations. Allows for Writing/Reading a client identity for a certain bank and
// Writing/Reading/Deleting coins related to a client.
type ClientStore struct {
	// db represents an active database connection. Used for creating transactions on each operation.
	db *sql.DB

	// clientId is the client's identity entry id on the database.
	clientId int64

	// BankName serves as the unique identifier for a bank.
	BankName string

	// LocalBalance keeps track of the local balance for this client.
	LocalBalance int64

	// RemoteBalance keeps track of the remote balance for this client.
	RemoteBalance int64
}

// BankStore handles a bank's local database operations. Allows for Writing/Reading a bank identity, Writing/Reading client's
// profiles and Writing/Reading deposits and exchanges information.
type BankStore struct {
	// db represents an active database connection. Used for creating transactions on each operation.
	db *sql.DB

	// Name is the Bank's public Name.
	Name string

	// identity serves as the unique identifier of a bank's identity.
	identity string
}
