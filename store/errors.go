package store

import "errors"

var (
	ErrExistingClient = errors.New("ziba/store: client already exists")
	ErrExistingCoin   = errors.New("ziba/store: coin already exists")
)
