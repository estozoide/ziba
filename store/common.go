package store

import (
	"database/sql"
	"log"
	"math/big"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Operation Type used for writing/deleting coins.
type Operation_Type int

const (
	Operation_Withdrawal Operation_Type = iota
	Operation_Payment
	Operation_Deposit
	Operation_Exchange
)

// GetZibaDir.
func GetZibaDir() (string, error) {
	// Get user's home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("failed to get home directory: %v", err)
		return "", err
	}

	// Set Ziba directory.
	ziba := filepath.Join(home, "Documents", "ziba-cli")

	// Create if don't exist.
	err = os.MkdirAll(ziba, 0755) // rwx r-x r-x
	if err != nil {
		log.Printf("failed to create Ziba directory: %v", err)
		return "", err
	}

	return ziba, nil
}

// openDatabase.
func openDatabase(dbPath string) (*sql.DB, error) {
	// Open database connection.
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Printf("failed to open database at %s: %v", dbPath, err)
		return nil, err
	}

	// Configure SQLite.
	pragmas := []string{
		"PRAGMA journal_mode=WAL",        // Enable WAL mode
		"PRAGMA busy_timeout=5000",       // Wait up to 5 seconds when database is locked
		"PRAGMA synchronous=NORMAL",      // Balance between safety and speed
		"PRAGMA cache_size=64000",        // 64MB cache size
		"PRAGMA foreign_keys=ON",         // Enable foreign key constraints
		"PRAGMA temp_store=MEMORY",       // Store temp tables and indices in memory
		"PRAGMA wal_autocheckpoint=1000", // Checkpoint WAL file every 1000 pages
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			log.Printf("failed to set pragma %s: %v", pragma, err)
			return nil, err
		}
	}

	return db, nil
}

// toString is used to translate big.Int types to string when writing to the database.
func toString(z *big.Int) string {
	if z == nil {
		return ""
	}
	return z.String()
}

// fromString is used to translate text scanned from the database into a big.Int type.
func fromString(s string) *big.Int {
	if s == "" {
		return nil
	}
	if z, ok := new(big.Int).SetString(s, 10); ok {
		return z
	}
	return nil
}

// rowScanner is a helper type for scanning rows from the database.
type rowScanner struct {
	dest []interface{}
}

// New allocates and returns a new rowScannner.
func (scanner *rowScanner) New(size int) *rowScanner {
	// Allocate an slice of empty interfaces.
	row := &rowScanner{
		dest: make([]interface{}, size),
	}

	// Make each element of the slice be a pointer to a string.
	for i := range row.dest {
		var s string
		row.dest[i] = &s
	}

	return row
}

// Strings returns the underlying string slice containing the column's values scanned from the database.
func (scanner *rowScanner) Strings() []string {
	// Allocate an slice of strings.
	res := make([]string, len(scanner.dest))

	// Grab the underlying value of each string pointer from rowScanner.
	for i, v := range scanner.dest {
		res[i] = *v.(*string)
	}

	return res
}
