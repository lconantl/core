package domain

import (
	"errors"
	"strings"
	"uuid"
)

var (
	ErrInvalidUserID       = errors.New("некорректный ID пользователя")
	ErrInvalidUsername     = errors.New("некорректное имя пользователя")
	ErrInvalidPasswordHash = errors.New("некорректный хэш пароля")
)

type Credentials struct {
	userID       uuid.UUID
	username     string
	passwordHash string
}

func NewCredentials(userID uuid.UUID, username string, passwordHash string) (*Credentials, error) {
	if userID == uuid.Nil() {
		return nil, ErrInvalidUserID
	}

	if strings.TrimSpace(username) == "" {
		return nil, ErrInvalidUsername
	}

	if passwordHash == "" {
		return nil, ErrInvalidPasswordHash
	}

	return &Credentials{
		userID:       userID,
		username:     username,
		passwordHash: passwordHash,
	}, nil
}

func (credentials *Credentials) UserID() uuid.UUID {
	return credentials.userID
}

func (credentials *Credentials) Username() string {
	return credentials.username
}

func (credentials *Credentials) PasswordHash() string {
	return credentials.passwordHash
}

func (credentials *Credentials) ChangePasswordHash(newHash string) error {
	if newHash == "" {
		return ErrInvalidPasswordHash
	}

	credentials.passwordHash = newHash
	return nil
}
