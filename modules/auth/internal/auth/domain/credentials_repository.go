package domain

import (
	"errors"
	"uuid"
)

var (
	ErrCredentialsNotFound = errors.New("учётные данные не найдены")
	ErrUsernameTaken       = errors.New("логин уже занят")
)

type CredentialsRepository interface {
	Add(credentials *Credentials) error
	Update(credentials *Credentials) error
	FindByUsername(username Username) (*Credentials, error)
	FindByUserIDForUpdate(userID uuid.UUID) (*Credentials, error)
}
