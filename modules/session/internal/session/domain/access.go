package domain

import (
	"errors"
	"time"
	"uuid"
)

var ErrInvalidAccessToken = errors.New("некорректный access-токен")

type AccessClaims struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type AccessTokenIssuer interface {
	Issue(claims AccessClaims) (string, error)
	Parse(token string, now time.Time) (AccessClaims, error)
}
