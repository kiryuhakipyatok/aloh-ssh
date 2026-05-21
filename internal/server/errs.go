package server

import (
	"aloh-ssh/pkg/errs"
	"errors"
)

const (
	SUCCESS = iota
	NOT_FOUND
	ALREADY_EXISTS
	SERVER_ERROR
)

func castErr(err error) []byte {
	if errors.Is(err, errs.ErrAlreadyExistsBase) {
		return []byte{ALREADY_EXISTS}
	}

	if errors.Is(err, errs.ErrNotFoundBase) {
		return []byte{NOT_FOUND}
	}

	return []byte{SERVER_ERROR}
}
