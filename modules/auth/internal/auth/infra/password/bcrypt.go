package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmptyPassword = errors.New("пароль должен быть заполнен")

func Hash(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(plain),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func Verify(plainPassword string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(plainPassword),
	)
}
