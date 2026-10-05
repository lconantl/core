package domain

import (
	"context"
	"errors"
)

var ErrCredentialsNotFound = errors.New("учётные данные не найдены")

type CredentialsRepository interface {
	FindByUsername(ctx context.Context, username string) (*Credentials, error)
}
