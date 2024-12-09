package core

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"log"
	"math/big"
	"time"
)

//
// SETUP (1/6)
//

// 1. A Central Authority selects a Sophie-Germain prime and its related safe prime. Computes a generator.
//	  This are the scheme parameters.
// 2. A Bank joins the scheme by creating an identity (from which its public identity can be computed).

// New allocates and returns a new SchemeParams.
func (scheme *SchemeParams) New() *SchemeParams {
	// Variables to set.
	var p, q, g *big.Int
	var err error

	// Find Sophie-Germain prime (q) and its related safe prime (p).
	for {
		// Generate a random prime number of length 1024 bits.
		q, err = rand.Prime(rand.Reader, 1024)
		if err != nil {
			log.Printf("failed to generate random number q")
			return nil
		}

		// Compute p = 2q + 1 and check if its a prime number.
		p = new(big.Int).Mul(q, big.NewInt(2))
		p.Add(p, big.NewInt(1))

		if ok := p.ProbablyPrime(20); ok {
			break
		}
	}

	// Find generator (g) in Z_p^*.
	g, err = rand.Prime(rand.Reader, 1024)
	if err != nil {
		return nil
	}

	// for {
	// 	h, err := rand.Prime(rand.Reader, 1024)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	// Test primitive element h by checking h^alpha != 1 mod p.
	// 	// Where alpha is { factors of p - 1 }.
	// 	// By p being p = 2q + 1 -> p - 1 = 2q.
	// 	// Factors of p - 1 are 2 and q.
	// 	h2 := new(big.Int).Exp(h, big.NewInt(2), p)
	// 	hq := new(big.Int).Exp(h, q, p)
	// 	if h2.Cmp(big.NewInt(1)) != 0 && hq.Cmp(big.NewInt(1)) != 0 {
	// 		g = h
	// 		break
	// 	}
	// }

	scheme.Q = q
	scheme.P = p
	scheme.G = g

	return scheme
}

// New allocates an returns a new RsaKey.
func (key *RsaKey) New() *RsaKey {
	// Generate RSA key of length 2048 bits.
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("failed to generate RSA key")
		return nil
	}

	key.P = rsaKey.Primes[0]
	key.Q = rsaKey.Primes[1]
	key.N = rsaKey.PublicKey.N
	key.D = rsaKey.D
	key.E = big.NewInt(int64(rsaKey.PublicKey.E))

	return key
}

// New allocates and returns a new Bank computed using scheme.
func (bank *Bank) New(scheme *SchemeParams) *Bank {
	// Check for valid SchemeParams.
	if scheme == nil {
		return nil
	}

	// Generate private identity number (x).
	priv, err := rand.Int(rand.Reader, scheme.P)
	if err != nil {
		log.Printf("failed to generate private identity number for Bank")
		return nil
	}

	// Compute public identity number (z).
	pub := new(big.Int).Exp(scheme.G, priv, scheme.P)

	// Generate RSA key.
	key := new(RsaKey).New()
	if key == nil {
		return nil
	}

	bank.Scheme = *scheme
	bank.Key = *key
	bank.Priv = priv
	bank.Pub = pub

	return bank
}

// Profile allocates and returns a new BankProfile using bank.
func (bank *Bank) Profile() *BankProfile {
	return &BankProfile{
		Scheme: bank.Scheme,
		Pub:    bank.Pub,
		N:      bank.Key.N,
		E:      bank.Key.E,
	}
}

//
// ACCOUNT GENERATION (2/6)
//

// 1. A Client joins the scheme by creating an identity associated to a certains bank's public identity
// 		(this client's identity can be used to calculate its public identity).
// 2. The Bank accepts the client's public identity and issues a credential and contract for this client.

// New allocates and returns a new Client computed using bank.
func (client *Client) New(bank *BankProfile) *Client {
	// Check for valid BankProfile.
	if bank == nil {
		return nil
	}

	// Generate private identity number (r_m).
	priv, err := rand.Int(rand.Reader, bank.Scheme.P)
	if err != nil {
		log.Printf("failed to generate private identity number for Client")
		return nil
	}

	// Generate public identity number (m).
	pub, err := rand.Int(rand.Reader, bank.N)
	if err != nil {
		log.Printf("failed to generate public identity number for Client")
		return nil
	}

	// Generate transaction identifier (ID_M).
	tradeId, err := rand.Int(rand.Reader, new(big.Int).Sub(bank.N, big.NewInt(1)))
	if err != nil {
		log.Printf("failed to generate transaction identifier for Client")
		return nil
	}

	// Generate RSA key.
	key := new(RsaKey).New()
	if key == nil {
		return nil
	}

	client.Bank = *bank
	client.Key = *key
	client.TradeId = tradeId
	client.Priv = priv
	client.Pub = pub

	return client
}

// Profile allocates and returns a new ClientProfile using client.
func (client *Client) Profile() *ClientProfile {
	// Compute private identity stamp number.
	privStamp := new(big.Int).Exp(client.Bank.Scheme.G, client.Priv, client.Bank.Scheme.P)

	// Compute identity hash number.
	identityHashBytes := sha256.Sum256(append(client.Pub.Bytes(), privStamp.Bytes()...))
	identityHash := new(big.Int).SetBytes(identityHashBytes[:])

	return &ClientProfile{
		PrivStamp:    privStamp,
		IdentityHash: identityHash,
		TradeId:      client.TradeId,
		Pub:          client.Pub,
		N:            client.Key.N,
		E:            client.Key.E,
	}
}

// concatenateBigInts allocated and returns a new Int computed like (first||second).
func concatenateBigInts(first, second *big.Int) *big.Int {
	secondBitLen := second.BitLen()
	result := new(big.Int)
	result.Lsh(first, uint(secondBitLen))
	result.Add(result, second)
	return result
}

// NewClient allocates and returns a new ClientInfo using profile.
func (bank *Bank) NewClient(profile *ClientProfile) (*ClientInfo, error) {
	// Verify client's identity.
	computedIdentityHashBytes := sha256.Sum256(append(profile.Pub.Bytes(), profile.PrivStamp.Bytes()...))
	computedIdentityHash := new(big.Int).SetBytes(computedIdentityHashBytes[:])
	if profile.IdentityHash.Cmp(computedIdentityHash) != 0 {
		return nil, ErrIdentityMismatch
	}

	// Generate randomizing number (k).
	k, err := rand.Int(rand.Reader, bank.Scheme.P)
	if err != nil {
		log.Printf("failed to generate random number")
		return nil, err
	}

	// Compute the blinded client's public identity number (s).
	s := new(big.Int).Mod(concatenateBigInts(profile.Pub, k), bank.Scheme.P)

	// Compute the client's credential (v).
	credential := new(big.Int).Exp(bank.Scheme.G, s, bank.Scheme.P)

	// Compute the client's contract (R).
	contract := new(big.Int).Exp(credential, bank.Priv, bank.Scheme.P)

	client := &ClientInfo{
		Profile:    *profile,
		K:          k,
		S:          s,
		Credential: credential,
		Contract:   contract,
	}

	return client, nil
}

// AddCredentials sets Credential, Contract for client and returns it.
func (client *Client) SetCredentials(credential *big.Int, contract *big.Int) *Client {
	client.Credential = credential
	client.Contract = contract
	return client
}

//
// WITHDRAWAL (3/6)
//

// 1. The Client generates a partial coin which consists of a number of parameters.
// 2. The Bank accepts this partial coin as a withdrawal request and computes some parameters that represent the Bank's
//		authorization of this coin.
// 3. The Client uses the Bank's issued parameters to compute some final coin parameters, therefore
//		completing the coin.

// random sets coin.Random to a new CoinRandom.
func (coin *Coin) random(client *Client) error {
	// Helper
	var err error

	// Generate random number (e).
	e, err := rand.Int(rand.Reader, client.Bank.Scheme.P)
	if err != nil {
		log.Printf("failed to generate random number")
		return err
	}

	// Generate random number (l) such that its inverse exists (l^-1).
	var l, lInv *big.Int
	for {
		l, err = rand.Int(rand.Reader, client.Bank.N)
		if err != nil {
			log.Printf("failed to generate random number")
			return err
		}

		lInv = new(big.Int).ModInverse(l, client.Bank.N)
		if lInv != nil {
			break
		}
	}

	// Generate random number (beta_1) such that its inverse exists (beta_1^-1).
	var beta1, beta1Inv *big.Int
	for {
		beta1, err = rand.Int(rand.Reader, client.Bank.Scheme.Q)
		if err != nil {
			log.Printf("failed to generate random number")
			return err
		}

		beta1Inv = new(big.Int).ModInverse(beta1, client.Bank.Scheme.Q)
		if beta1Inv != nil {
			break
		}
	}

	// Generate random number (y) such that its inverse exists (y^-1).
	var y, yInv *big.Int
	pMinus1 := new(big.Int).Sub(client.Bank.Scheme.P, big.NewInt(1))
	for {
		y, err = rand.Int(rand.Reader, pMinus1)
		if err != nil {
			log.Printf("failed to generate random number")
			return err
		}

		yInv = new(big.Int).ModInverse(y, pMinus1)
		if yInv != nil {
			break
		}
	}

	// Generate random number (beta_2).
	beta2, err := rand.Int(rand.Reader, client.Bank.Scheme.P)
	if err != nil {
		log.Printf("failed to generate random number")
		return err
	}

	coin.Random = CoinRandom{
		E:        e,
		L:        l,
		LInv:     lInv,
		Beta1:    beta1,
		Beta1Inv: beta1Inv,
		Beta2:    beta2,
		Y:        y,
		YInv:     yInv,
	}

	return nil
}

// elgamal sets  coin.Elgamal to a new CoinElgamal.
func (coin *Coin) elgamal(client *Client) {
	// Compute Elgamal private key (w).
	priv := new(big.Int).Mod(concatenateBigInts(client.Contract, coin.Random.E), client.Bank.Scheme.P)

	// Compute Elgamal public key (alpha).
	pub := new(big.Int).Exp(client.Bank.Scheme.G, priv, client.Bank.Scheme.P)

	// Compute Elgamal first component (u).
	first := new(big.Int).Exp(client.Bank.Scheme.G, coin.Random.Y, client.Bank.Scheme.P)

	coin.Elgamal = CoinElgamal{
		Priv:  priv,
		Pub:   pub,
		First: first,
	}
}

// params sets coin.Params to a new CoinParams.
func (coin *Coin) params(client *Client) {
	// Compute client's blinded credential (A).
	A := new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).Exp(client.Credential, coin.Random.Beta1, client.Bank.Scheme.P),
			new(big.Int).Exp(client.Bank.Scheme.G, coin.Random.Beta2, client.Bank.Scheme.P),
		),
		client.Bank.Scheme.P,
	)

	// Compute blind signature envelope for A (a).
	a := new(big.Int).Mod(
		new(big.Int).Mul(
			A,
			new(big.Int).Exp(coin.Random.L, client.Bank.E, client.Bank.N),
		),
		client.Bank.N,
	)

	// Compute digest of some coin parameters.
	var buffer bytes.Buffer
	buffer.Write(coin.Elgamal.First.Bytes())
	buffer.Write(coin.Elgamal.Pub.Bytes())
	buffer.Write(A.Bytes())
	hashBytes := sha256.Sum256(buffer.Bytes())
	hash := new(big.Int).SetBytes(hashBytes[:])

	// Compute signature envelope for some coin parameters (C).
	C := new(big.Int).Mod(
		new(big.Int).Mul(coin.Random.Beta1Inv, hash),
		client.Bank.Scheme.Q,
	)

	coin.Params = CoinParams{
		A:      A,
		ALower: a,
		C:      C,
	}
}

// NewCoinRequest generates a partial coin to be used for a withdrawal request.
func (client *Client) NewCoinRequest() *Coin {
	// Empty Coin object.
	coin := new(Coin)

	// Fill Coin.Random.
	err := coin.random(client)
	if err != nil {
		return nil
	}

	// Fill Coin.Elgamal.
	coin.elgamal(client)

	// Fill Coin.Params.
	coin.params(client)

	return coin
}

// NewCoinResponse computes some of the final coin parameters as a withdrawal response.
func (bank *Bank) NewCoinResponse(client *ClientInfo, ALower *big.Int, C *big.Int) (Expiration time.Time, A1 *big.Int, C1 *big.Int) {
	// Choose an expiration date for the coin (t). In this case is one month and one day from the current time.
	Expiration = time.Now().AddDate(0, 1, 1)
	expirationBytes, _ := Expiration.MarshalBinary()

	// Compute digest of expiration date.
	hashBytes := sha256.Sum256(expirationBytes)
	hash := new(big.Int).SetBytes(hashBytes[:])

	// Compute a blind signature on A (A').
	A1 = new(big.Int).Exp(
		new(big.Int).Mul(ALower, hash),
		bank.Key.D,
		bank.Key.N,
	)

	// Compute a signature on c (c').
	C1 = new(big.Int).Mod(
		new(big.Int).Add(
			new(big.Int).Mul(C, bank.Priv),
			client.S,
		),
		bank.Scheme.Q,
	)

	return
}

// FinishCoin computes a complete coin using the bank's reponse.
func (client *Client) FinishCoin(coin *Coin, Expiration time.Time, A1 *big.Int, C1 *big.Int) *Coin {
	// Reveal the blind signature on  A (A'').
	A2 := new(big.Int).Mod(
		new(big.Int).Mul(coin.Random.LInv, A1),
		client.Bank.N,
	)

	// Reveal the signature on c (R).
	R := new(big.Int).Mod(
		new(big.Int).Add(
			new(big.Int).Mul(coin.Random.Beta1, C1),
			coin.Random.Beta2,
		),
		client.Bank.Scheme.Q,
	)

	coin.Params.A1 = A1
	coin.Params.C1 = C1
	coin.Params.Expiration = Expiration
	coin.Params.A2 = A2
	coin.Params.R = R

	return coin
}

// Profile allocates and returns a new CoinProfile from coin.
func (coin *Coin) Profile() *CoinProfile {
	return &CoinProfile{
		Pub:        coin.Elgamal.Pub,
		First:      coin.Elgamal.First,
		A:          coin.Params.A,
		R:          coin.Params.R,
		A2:         coin.Params.A2,
		Expiration: coin.Params.Expiration,
		Second:     coin.Elgamal.Second,
		Msg:        coin.Elgamal.Msg,
	}
}

//
// PAYMENT (4/6)
//

// 1. A Client (here known as the Spender) sends a coin (its profile) to another Client (known as Merchant).
// 2. The Merchant verifies the coin properties. And creates the Elgamal's message.
// 3. The Spender signs the coin using the Merchant's message (using Elgamal).
// 4. The Merchant verifies the Elgamal's signature on the message.

// VerifyProperties verifies both of the Coin's properties and returns a success bool.
func (coin *CoinProfile) VerifyProperties(bank *BankProfile) bool {
	// Compute digest of expiration date.
	expirationBytes, _ := coin.Expiration.MarshalBinary()
	hashBytes := sha256.Sum256(expirationBytes)
	hash := new(big.Int).SetBytes(hashBytes[:])

	// Compute left-side of first property.
	left := new(big.Int).Mod(new(big.Int).Mul(coin.A, hash), bank.N)

	// Compute right-side of first property.
	right := new(big.Int).Exp(coin.A2, bank.E, bank.N)

	// Verify first property.
	if left.Cmp(right) != 0 {
		return false
	}

	// Compute left-side of second property.
	left = new(big.Int).Exp(bank.Scheme.G, coin.R, bank.Scheme.P)

	// Compute digest of some coin parameters.
	var buffer bytes.Buffer
	buffer.Write(coin.First.Bytes())
	buffer.Write(coin.Pub.Bytes())
	buffer.Write(coin.A.Bytes())
	hashBytes = sha256.Sum256(buffer.Bytes())
	hash = new(big.Int).SetBytes(hashBytes[:])

	// Compute right-side of second property.
	right = new(big.Int).Mod(
		new(big.Int).Mul(
			coin.A,
			new(big.Int).Exp(bank.Pub, hash, bank.Scheme.P),
		),
		bank.Scheme.P,
	)

	return left.Cmp(right) == 0
}

// Stamp computes the Elgamal's message using some transaction parameters and returns it.
func (coin *CoinProfile) Stamp(bank *BankProfile, client *ClientProfile) (msg *big.Int) {
	// Compute the current time as the transaction date (t).
	t := time.Now()
	tBytes, _ := t.MarshalBinary()

	// Compute the hash of some coin parameters.
	var buffer bytes.Buffer
	buffer.Write(coin.Pub.Bytes())
	buffer.Write(coin.First.Bytes())
	buffer.Write(client.TradeId.Bytes())
	buffer.Write(tBytes)

	// Compute the Elgamal message as the digest of the coin parameters (d).
	hashBytes := sha256.Sum256(buffer.Bytes())
	msg = new(big.Int).SetBytes(hashBytes[:])

	coin.Msg = msg

	return
}

// SignCoin computes the Elgamal's second component using the message and returns it.
func (client *Client) SignCoin(coin *Coin, msg *big.Int) (second *big.Int) {
	// Set msg on coin.
	coin.Elgamal.Msg = msg

	// Helper.
	pMinus1 := new(big.Int).Sub(client.Bank.Scheme.P, big.NewInt(1))

	// Compute Elgamal's second component (gamma).
	second = new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).Sub(
				coin.Elgamal.Msg,
				new(big.Int).Mul(coin.Elgamal.Priv, coin.Elgamal.First),
			),
			coin.Random.YInv,
		),
		pMinus1,
	)

	coin.Elgamal.Second = second

	return
}

// VerifyElgamal verifies the Elgamal's identity and returns a success bool.
func (coin *CoinProfile) VerifyElgamal(bank *BankProfile, second *big.Int) bool {
	// Set second on coin.
	coin.Second = second

	// Compute left-side of Elgamal's identity.
	left := new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).Exp(coin.Pub, coin.First, bank.Scheme.P),
			new(big.Int).Exp(coin.First, coin.Second, bank.Scheme.P),
		),
		bank.Scheme.P,
	)

	// Compute right-side of Elgamal's identity.
	right := new(big.Int).Exp(bank.Scheme.G, coin.Msg, bank.Scheme.P)

	return left.Cmp(right) == 0
}
