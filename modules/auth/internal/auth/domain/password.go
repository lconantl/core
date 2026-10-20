package domain

import (
	"errors"
	"unicode/utf8"
)

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72
)

var (
	ErrPasswordTooShort = errors.New("пароль слишком короткий")
	ErrPasswordTooLong  = errors.New("пароль слишком длинный")
	ErrWrongPassword    = errors.New("неверный пароль")
)

type Password struct {
	value string
}

func NewPassword(value string) (Password, error) {
	if utf8.RuneCountInString(value) < minPasswordLength {
		return Password{}, ErrPasswordTooShort
	}

	if len(value) > maxPasswordBytes {
		return Password{}, ErrPasswordTooLong
	}

	return Password{value: value}, nil
}

func (password Password) Plain() string {
	return password.value
}

func (password Password) String() string {
	return "[hidden]"
}

func (password Password) IsZero() bool {
	return password.value == ""
}

type PasswordHasher interface {
	Hash(password Password) (string, error)
	Verify(plain string, hash string) error
}
