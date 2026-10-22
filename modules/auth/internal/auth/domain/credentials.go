package domain

import (
	"errors"
	"uuid"
)

var (
	ErrInvalidUserID       = errors.New("некорректный ID пользователя")
	ErrInvalidPassword     = errors.New("пароль не задан")
	ErrInvalidPasswordHash = errors.New("некорректный хэш пароля")
)

type Credentials struct {
	userID       uuid.UUID
	username     Username
	passwordHash string
}

func NewCredentials(userID uuid.UUID, username Username, password Password, hasher PasswordHasher) (*Credentials, error) {
	err := validateIdentity(userID, username)

	if err != nil {
		return nil, err
	}

	if password.IsZero() {
		return nil, ErrInvalidPassword
	}

	hash, err := hasher.Hash(password)

	if err != nil {
		return nil, err
	}

	return LoadCredentials(userID, username, hash)
}

func LoadCredentials(userID uuid.UUID, username Username, passwordHash string) (*Credentials, error) {
	err := validateIdentity(userID, username)

	if err != nil {
		return nil, err
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

func validateIdentity(userID uuid.UUID, username Username) error {
	if userID == uuid.Nil() {
		return ErrInvalidUserID
	}

	if username.IsZero() {
		return ErrInvalidUsername
	}

	return nil
}

func (credentials *Credentials) UserID() uuid.UUID {
	return credentials.userID
}

func (credentials *Credentials) Username() Username {
	return credentials.username
}

func (credentials *Credentials) PasswordHash() string {
	return credentials.passwordHash
}

func (credentials *Credentials) VerifyPassword(plain string, hasher PasswordHasher) error {
	return hasher.Verify(plain, credentials.passwordHash)
}

func (credentials *Credentials) ChangePassword(oldPlain string, newPassword Password, hasher PasswordHasher) error {
	err := credentials.VerifyPassword(oldPlain, hasher)

	if err != nil {
		return err
	}

	if newPassword.IsZero() {
		return ErrInvalidPassword
	}

	hash, err := hasher.Hash(newPassword)

	if err != nil {
		return err
	}

	credentials.passwordHash = hash

	return nil
}
