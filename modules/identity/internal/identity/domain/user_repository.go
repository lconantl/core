package domain

import "errors"

var (
	ErrUserNotFound = errors.New("пользователь не найден")
	ErrEmailTaken   = errors.New("почта уже занята")
)

type UserRepository interface {
	Add(user *User) error
	Update(user *User) error
	Find(id string) (*User, error)
	FindByIDForUpdate(id string) (*User, error)
}
