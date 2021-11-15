package vmess

import (
	"errors"
)

var (
	ErrNotFound = errors.New("user do not exist")
	ErrReplay   = errors.New("replayed request")
)
