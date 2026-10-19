package domain

import (
	"errors"
	"uuid"
)

var (
	ErrInvalidUserID           = errors.New("некорректный ID пользователя")
	ErrInvalidUserStatus       = errors.New("некорректный статус пользователя")
	ErrInvalidStatusTransition = errors.New("некорректная смена статуса")
)

const (
	UserStatusPending UserStatus = "pending"
	UserStatusActive  UserStatus = "active"
	UserStatusDeleted UserStatus = "deleted"
)

type UserStatus string

type User struct {
	id      uuid.UUID
	status  UserStatus
	profile Profile
}

func NewUser(id uuid.UUID, profile Profile) (*User, error) {
	return LoadUser(id, UserStatusPending, profile)
}

func LoadUser(id uuid.UUID, status UserStatus, profile Profile) (*User, error) {
	if id == uuid.Nil() {
		return nil, ErrInvalidUserID
	}

	if !status.IsValid() {
		return nil, ErrInvalidUserStatus
	}

	if profile.IsZero() {
		return nil, ErrInvalidProfile
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

func (user *User) UpdateProfile(profile Profile) error {
	if profile.IsZero() {
		return ErrInvalidProfile
	}

	user.profile = profile
	return nil
}

func (user *User) Activate() error {
	switch user.status {
	case UserStatusPending:
		user.status = UserStatusActive
		return nil

	case UserStatusActive:
		return nil

	default:
		return ErrInvalidStatusTransition
	}
}

func (user *User) Delete() {
	user.status = UserStatusDeleted
}

func (status UserStatus) IsValid() bool {
	switch status {
	case UserStatusPending, UserStatusActive, UserStatusDeleted:
		return true
	}

	return false
}
