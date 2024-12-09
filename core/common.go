package core

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"io"
	"log"
	"os"
)

//go:embed params.json
var files embed.FS

// Params.
var Params *SchemeParams

// init.
func init() {
	// Open params file.
	paramsFile, err := files.Open("params.json")
	if err != nil {
		log.Fatalf("failed to load params.json: %v", err)
	}

	// Load into variable.
	scheme := new(SchemeParams)
	err = LoadFromFile(scheme, paramsFile)
	if err != nil {
		log.Fatalf("failed to load SchemeParams from file: %v", err)
	}

	Params = scheme
}

// Hash computes the digest of the contents of coin and returns a truncated result.
func (coin *CoinProfile) Hash() uint32 {
	// Date to bytes.
	expirationBytes, _ := coin.Expiration.MarshalBinary()

	// Helper byte buffer.
	var buffer bytes.Buffer
	buffer.Write(coin.Pub.Bytes())
	buffer.Write(coin.First.Bytes())
	buffer.Write(coin.A.Bytes())
	buffer.Write(coin.R.Bytes())
	buffer.Write(coin.A2.Bytes())
	buffer.Write(expirationBytes)

	// Actually compute the digest from the buffer.
	hashBytes := sha256.Sum256(buffer.Bytes())

	// Truncate the result to fit into an int64.
	hash := int64(hashBytes[0]) | int64(hashBytes[1])<<8 | int64(hashBytes[2])<<16 |
		int64(hashBytes[3])<<24 | int64(hashBytes[4])<<32 | int64(hashBytes[5])<<40 |
		int64(hashBytes[6])<<48 | int64(hashBytes[7])<<56

	return uint32(hash)
}

// Hash computes the digest of the contents of client and returns a truncated result.
func (client *ClientProfile) Hash() uint32 {
	// Helper byte buffer.
	var buffer bytes.Buffer
	buffer.Write(client.PrivStamp.Bytes())
	buffer.Write(client.IdentityHash.Bytes())
	buffer.Write(client.TradeId.Bytes())
	buffer.Write(client.Pub.Bytes())
	buffer.Write(client.N.Bytes())
	buffer.Write(client.E.Bytes())

	// Actually compute the digest from the buffer.
	hashBytes := sha256.Sum256(buffer.Bytes())

	// Truncate the result to fit into an int64.
	hash := int64(hashBytes[0]) | int64(hashBytes[1])<<8 | int64(hashBytes[2])<<16 |
		int64(hashBytes[3])<<24 | int64(hashBytes[4])<<32 | int64(hashBytes[5])<<40 |
		int64(hashBytes[6])<<48 | int64(hashBytes[7])<<56

	return uint32(hash)
}

// Save to .json.
func SaveToFile(data json.Marshaler, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("failed to create file")
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Load from .json.
func LoadFromFile(target json.Unmarshaler, file io.ReadCloser) error {
	// file, err := os.Open(filename)
	// if err != nil {
	// 	log.Printf("failed to open file %s", filename)
	// 	return err
	// }
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(target)
}
