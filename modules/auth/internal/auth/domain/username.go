package domain

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidUsername = errors.New("некорректное имя пользователя")

var usernamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,31}$`)

type Username struct {
	value string
}

func NewUsername(value string) (Username, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	if !usernamePattern.MatchString(value) {
		return Username{}, ErrInvalidUsername
	}

	return Username{value: value}, nil
}

func (username Username) String() string {
	return username.value
}

func (username Username) IsZero() bool {
	return username.value == ""
}
