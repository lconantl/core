package domain

import (
	"errors"
	"uuid"
)

var (
	ErrorInvalidUserID     = errors.New("некорректный ID пользователя")
	ErrorInvalidUserStatus = errors.New("некорректный статус пользователя")
)

const (
	UserStatusPending UserStatus = "pending"
	UserStatusActive  UserStatus = "active"
)

type UserStatus string

type User struct {
	id      uuid.UUID
	status  UserStatus
	profile Profile
}

func NewUser(id uuid.UUID, profile Profile) (*User, error) {
	return RestoreUser(id, UserStatusPending, profile)
}

func RestoreUser(id uuid.UUID, status UserStatus, profile Profile) (*User, error) {
	if id == uuid.Nil() {
		return nil, ErrorInvalidUserID
	}

	switch status {
	case UserStatusPending, UserStatusActive:
	default:
		return nil, ErrorInvalidUserStatus
	}

	return &User{
		id:      id,
		status:  status,
		profile: profile,
	}, nil
}

func (user *User) ID() uuid.UUID {
	return user.id
}

func (user *User) Status() UserStatus {
	return user.status
}

func (user *User) Profile() Profile {
	return user.profile
}

func (user *User) UpdateProfile(profile Profile) {
	user.profile = profile
}

func (user *User) Activate() error {
	switch user.status {
	case UserStatusPending:
		user.status = UserStatusActive
		return nil

	case UserStatusActive:
		return nil

	default:
		return ErrorInvalidUserStatus
	}
}
