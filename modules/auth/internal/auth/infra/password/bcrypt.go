package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"auth/internal/auth/domain"
)

var _ domain.PasswordHasher = (*Bcrypt)(nil)

type Bcrypt struct {
	cost int
}

func NewBcrypt(cost int) *Bcrypt {
	return &Bcrypt{cost: cost}
}

func (hasher *Bcrypt) Hash(password domain.Password) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password.Plain()), hasher.cost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (hasher *Bcrypt) Verify(plain string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return domain.ErrWrongPassword
	}

	return err
}
