package errs

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnavailable     = errors.New("unavailable")
)

func Unauthenticated(err error) error {
	if err == nil {
		return ErrUnauthenticated
	}

	return fmt.Errorf("%w: %w", ErrUnauthenticated, err)
}

func Unavailable(err error) error {
	if err == nil {
		return ErrUnavailable
	}

	return fmt.Errorf("%w: %w", ErrUnavailable, err)
}

func IsUnauthenticated(err error) bool {
	return errors.Is(err, ErrUnauthenticated)
}

func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

func UnauthenticatedIfNoRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return Unauthenticated(err)
	}

	return err
}

func UnavailableIfNoRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return Unavailable(err)
	}

	return err
}
