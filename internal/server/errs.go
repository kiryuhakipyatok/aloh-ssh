package server

import (
	"aloh-ssh/pkg/errs"
	"errors"
	"log"
)

const (
	SUCCESS = iota
	NOT_FOUND
	ALREADY_EXISTS
	SERVER_ERROR
)

func castErr(err error) []byte {
	res := make([]byte, 0, 1)
	apperr, ok := errors.AsType[errs.AppError](err)
	if ok {
		switch apperr {
		case errs.ErrAlreadyExistsBase:
			res = append(res, ALREADY_EXISTS)
		case errs.ErrNotFoundBase:
			res = append(res, NOT_FOUND)
		default:
			res = append(res, SERVER_ERROR)
		}
	} else {
		res = append(res, SERVER_ERROR)
	}
	log.Println(string(res))
	return res
}
