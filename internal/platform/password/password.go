// Provides password hashing and verification for teacher and student login.

package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinLength = 7
	MaxLength = 128
)

var ErrInvalidLength = errors.New("password length must be between 7 and 128")

func ValidatePlaintext(value string) error {
	if len(value) < MinLength || len(value) > MaxLength {
		return ErrInvalidLength
	}

	return nil
}

func Hash(value string) (string, error) {
	if err := ValidatePlaintext(value); err != nil {
		return "", err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func Verify(hash, value string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)) == nil
}
