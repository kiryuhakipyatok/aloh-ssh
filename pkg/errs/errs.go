package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNotFoundBase                 = errors.New("not found")
	ErrAlreadyExistsBase            = errors.New("already exists")
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
	return NewAppError(op, fmt.Errorf("%w : %w", ErrAlreadyExistsBase, err))
}

func ErrNotFound(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrNotFoundBase))
}