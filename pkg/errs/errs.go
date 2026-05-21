package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNotFoundBase       = errors.New("not found")
	ErrAlreadyExistsBase  = errors.New("already exists")
	ErrRequestTimeoutBase = errors.New("request timeout")
	ErrInvalidTypeBase    = errors.New("invalid type")
)

type AppError struct {
	Operation string
	Err       error
}

func (ae AppError) Error() string {
	return fmt.Sprintf("%s : %v", ae.Operation, ae.Err)
}

func (ae AppError) Unwrap() error {
	return ae.Err
}

func NewAppError(op string, err error) AppError {
	return AppError{
		Operation: op,
		Err:       err,
	}
}

func ErrAlreadyExists(op string, err error) AppError {
	return NewAppError(op, fmt.Errorf("%w: %w", ErrAlreadyExistsBase, err))
}

func ErrNotFound(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrNotFoundBase))
}

func ErrRequestTimeout(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrRequestTimeoutBase))
}

func ErrInvalidType(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrInvalidTypeBase))
}
