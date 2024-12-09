package core

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

//
// Helper formatting functions.
//

// formatBigInt formats a big.Int by showing only the first n digits.
func formatBigInt(n *big.Int, digits int) string {
	if n == nil {
		return "<nil>"
	}
	str := n.String()
	if len(str) > digits {
		return str[:digits] + "..."
	}
	return str
}

//
// String methods for all types.
//

// String satisfies the fmt.Stringer interface for SchemeParams.
func (scheme SchemeParams) String() string {
	var b strings.Builder
	b.WriteString("SchemeParams {\n")
	b.WriteString(fmt.Sprintf("# Q: %s\n", formatBigInt(scheme.Q, 100)))
	b.WriteString(fmt.Sprintf("# P: %s\n", formatBigInt(scheme.P, 100)))
	b.WriteString(fmt.Sprintf("# G: %s\n", formatBigInt(scheme.G, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for RsaKey.
func (key RsaKey) String() string {
	var b strings.Builder
	b.WriteString("RsaKey{\n")
	b.WriteString(fmt.Sprintf("# P: %s\n", formatBigInt(key.P, 100)))
	b.WriteString(fmt.Sprintf("# Q: %s\n", formatBigInt(key.Q, 100)))
	b.WriteString(fmt.Sprintf("# D: %s\n", formatBigInt(key.D, 100)))
	b.WriteString(fmt.Sprintf("# N: %s\n", formatBigInt(key.N, 100)))
	b.WriteString(fmt.Sprintf("# E: %s\n", formatBigInt(key.E, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for Bank.
func (bank Bank) String() string {
	var b strings.Builder
	b.WriteString("Bank {\n")
	b.WriteString(bank.Scheme.String())
	b.WriteString(bank.Key.String())
	b.WriteString(fmt.Sprintf("# Priv: %s\n", formatBigInt(bank.Priv, 100)))
	b.WriteString(fmt.Sprintf("# Pub:  %s\n", formatBigInt(bank.Pub, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for BankProfile.
func (profile BankProfile) String() string {
	var b strings.Builder
	b.WriteString("BankProfile {\n")
	b.WriteString(profile.Scheme.String())
	b.WriteString(fmt.Sprintf("# Pub: %s\n", formatBigInt(profile.Pub, 100)))
	b.WriteString(fmt.Sprintf("# N:   %s\n", formatBigInt(profile.N, 100)))
	b.WriteString(fmt.Sprintf("# E:   %s\n", formatBigInt(profile.E, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for Client.
func (client Client) String() string {
	var b strings.Builder
	b.WriteString("Client {\n")
	b.WriteString(client.Bank.String())
	b.WriteString(client.Key.String())
	b.WriteString(fmt.Sprintf("# TradeId:    %s\n", formatBigInt(client.TradeId, 100)))
	b.WriteString(fmt.Sprintf("# Priv:       %s\n", formatBigInt(client.Priv, 100)))
	b.WriteString(fmt.Sprintf("# Pub:        %s\n", formatBigInt(client.Pub, 100)))
	b.WriteString(fmt.Sprintf("# Credential: %s\n", formatBigInt(client.Credential, 100)))
	b.WriteString(fmt.Sprintf("# Contract:   %s\n", formatBigInt(client.Contract, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for ClientProfile.
func (profile ClientProfile) String() string {
	var b strings.Builder
	b.WriteString("ClientProfile {\n")
	b.WriteString(fmt.Sprintf("# PrivStamp:    %s\n", formatBigInt(profile.PrivStamp, 100)))
	b.WriteString(fmt.Sprintf("# IdentityHash: %s\n", formatBigInt(profile.IdentityHash, 100)))
	b.WriteString(fmt.Sprintf("# TradeId:      %s\n", formatBigInt(profile.TradeId, 100)))
	b.WriteString(fmt.Sprintf("# Pub:          %s\n", formatBigInt(profile.Pub, 100)))
	b.WriteString(fmt.Sprintf("# N:            %s\n", formatBigInt(profile.N, 100)))
	b.WriteString(fmt.Sprintf("# E:            %s\n", formatBigInt(profile.E, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for ClientInfo.
func (client ClientInfo) String() string {
	var b strings.Builder
	b.WriteString("ClientInfo {\n")
	b.WriteString(client.Profile.String())
	b.WriteString(fmt.Sprintf("# K:          %s\n", formatBigInt(client.K, 100)))
	b.WriteString(fmt.Sprintf("# S:          %s\n", formatBigInt(client.S, 100)))
	b.WriteString(fmt.Sprintf("# Credential: %s\n", formatBigInt(client.Credential, 100)))
	b.WriteString(fmt.Sprintf("# Contract:   %s\n", formatBigInt(client.Contract, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for CoinRandom.
func (random CoinRandom) String() string {
	var b strings.Builder
	b.WriteString("CoinRandom {\n")
	b.WriteString(fmt.Sprintf("# E:        %s\n", formatBigInt(random.E, 100)))
	b.WriteString(fmt.Sprintf("# L:        %s\n", formatBigInt(random.L, 100)))
	b.WriteString(fmt.Sprintf("# LInv:     %s\n", formatBigInt(random.LInv, 100)))
	b.WriteString(fmt.Sprintf("# Beta1:    %s\n", formatBigInt(random.Beta1, 100)))
	b.WriteString(fmt.Sprintf("# Beta1Inv: %s\n", formatBigInt(random.Beta1Inv, 100)))
	b.WriteString(fmt.Sprintf("# Beta2:    %s\n", formatBigInt(random.Beta2, 100)))
	b.WriteString(fmt.Sprintf("# Y:        %s\n", formatBigInt(random.Y, 100)))
	b.WriteString(fmt.Sprintf("# YInv:     %s\n", formatBigInt(random.YInv, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for CoinElgamal.
func (elgamal CoinElgamal) String() string {
	var b strings.Builder
	b.WriteString("CoinElgamal {\n")
	b.WriteString(fmt.Sprintf("# Priv:   %s\n", formatBigInt(elgamal.Priv, 100)))
	b.WriteString(fmt.Sprintf("# Pub:    %s\n", formatBigInt(elgamal.Pub, 100)))
	b.WriteString(fmt.Sprintf("# First:  %s\n", formatBigInt(elgamal.First, 100)))
	b.WriteString(fmt.Sprintf("# Second: %s\n", formatBigInt(elgamal.Second, 100)))
	b.WriteString(fmt.Sprintf("# Msg:    %s\n", formatBigInt(elgamal.Msg, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for CoinParams.
func (params CoinParams) String() string {
	var b strings.Builder
	b.WriteString("CoinParams {\n")
	b.WriteString(fmt.Sprintf("# A:          %s\n", formatBigInt(params.A, 100)))
	b.WriteString(fmt.Sprintf("# ALower:     %s\n", formatBigInt(params.ALower, 100)))
	b.WriteString(fmt.Sprintf("# C:          %s\n", formatBigInt(params.C, 100)))
	b.WriteString(fmt.Sprintf("# A1:         %s\n", formatBigInt(params.A1, 100)))
	b.WriteString(fmt.Sprintf("# C1:         %s\n", formatBigInt(params.C1, 100)))
	b.WriteString(fmt.Sprintf("# Expiration: %s\n", params.Expiration))
	b.WriteString(fmt.Sprintf("# R:          %s\n", formatBigInt(params.R, 100)))
	b.WriteString(fmt.Sprintf("# A2:         %s\n", formatBigInt(params.A2, 100)))
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for Coin.
func (coin Coin) String() string {
	var b strings.Builder
	b.WriteString("Coin {\n")
	b.WriteString(coin.Random.String())
	b.WriteString(coin.Elgamal.String())
	b.WriteString(coin.Params.String())
	b.WriteString("}\n")
	return b.String()
}

// String satisfies the fmt.Stringer interface for CoinProfile.
func (profile CoinProfile) String() string {
	var b strings.Builder
	b.WriteString("CoinProfile {\n")
	b.WriteString(fmt.Sprintf("# Pub:        %s\n", formatBigInt(profile.Pub, 100)))
	b.WriteString(fmt.Sprintf("# First:      %s\n", formatBigInt(profile.First, 100)))
	b.WriteString(fmt.Sprintf("# A:          %s\n", formatBigInt(profile.A, 100)))
	b.WriteString(fmt.Sprintf("# R:          %s\n", formatBigInt(profile.R, 100)))
	b.WriteString(fmt.Sprintf("# A2:         %s\n", formatBigInt(profile.A2, 100)))
	b.WriteString(fmt.Sprintf("# Expiration: %s\n", profile.Expiration))
	b.WriteString(fmt.Sprintf("# Second:     %s\n", formatBigInt(profile.Second, 100)))
	b.WriteString(fmt.Sprintf("# Msg:        %s\n", formatBigInt(profile.Msg, 100)))
	b.WriteString("}\n")
	return b.String()
}

//
// JSON encoder/decoder for some types.
//

// schemeParamsJSON represents the JSON-friendly structure for SchemeParams.
type schemeParamsJSON struct {
	Q string `json:"Q"`
	P string `json:"P"`
	G string `json:"G"`
}

// MarshalJSON converts SchemeParams to JSON format.
func (s *SchemeParams) MarshalJSON() ([]byte, error) {
	wrapper := schemeParamsJSON{
		Q: s.Q.String(),
		P: s.P.String(),
		G: s.G.String(),
	}
	return json.Marshal(wrapper)
}

// UnmarshalJSON populates SchemeParams from JSON data.
func (s *SchemeParams) UnmarshalJSON(data []byte) error {
	var wrapper schemeParamsJSON
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	s.Q, _ = new(big.Int).SetString(wrapper.Q, 10)
	s.P, _ = new(big.Int).SetString(wrapper.P, 10)
	s.G, _ = new(big.Int).SetString(wrapper.G, 10)
	return nil
}
