package password_test

import (
	"auth/internal/auth/domain"
	"auth/internal/auth/infra/password"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newHasher() *password.Bcrypt {
	return password.NewBcrypt(bcrypt.MinCost)
}

func mustPassword(test *testing.T, raw string) domain.Password {
	test.Helper()

	value, err := domain.NewPassword(raw)
	require.NoError(test, err)

	return value
}

func TestBcryptHashAndVerify(test *testing.T) {
	hasher := newHasher()

	hash, err := hasher.Hash(mustPassword(test, "correct-horse"))
	require.NoError(test, err)

	require.NotEqual(test, "correct-horse", hash)
	require.NoError(test, hasher.Verify("correct-horse", hash))
}

func TestBcryptWrongPassword(test *testing.T) {
	hasher := newHasher()

	hash, err := hasher.Hash(mustPassword(test, "correct-horse"))
	require.NoError(test, err)

	require.ErrorIs(test, hasher.Verify("wrong-horse", hash), domain.ErrWrongPassword)
}

func TestBcryptSaltedHashes(test *testing.T) {
	hasher := newHasher()
	plain := mustPassword(test, "correct-horse")

	first, err := hasher.Hash(plain)
	require.NoError(test, err)
	second, err := hasher.Hash(plain)
	require.NoError(test, err)

	require.NotEqual(test, first, second, "у одинаковых паролей хеши должны отличаться из-за соли")
}

func TestBcryptBrokenHash(test *testing.T) {
	err := newHasher().Verify("correct-horse", "not-a-bcrypt-hash")

	require.Error(test, err)
	require.NotErrorIs(test, err, domain.ErrWrongPassword, "битый хеш — это сбой, а не неверный пароль")
}
