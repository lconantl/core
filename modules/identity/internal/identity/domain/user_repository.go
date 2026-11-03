package domain

import (
	"errors"
	"uuid"
)

var (
	ErrUserNotFound = errors.New("пользователь не найден")
	ErrEmailTaken   = errors.New("почта уже занята")
)

type UserRepository interface {
	Add(user *User) error
	Update(user *User) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*User, error)
	FindByIDForUpdate(id uuid.UUID) (*User, error)
}
