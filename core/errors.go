package core

import (
	"errors"
)

var (
	ErrIdentityMismatch = errors.New("ziba/core: verification error at IdentityHash")
)
