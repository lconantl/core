package integration_tests

import (
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"auth/internal/auth/domain"
	"auth/internal/auth/infra/mysql/repository"
)

func credentialsRepository(t *testing.T) domain.CredentialsRepository {
	t.Helper()
	return repository.NewCredentialsRepository(t.Context(), testEnv.connector.TransactionalClient())
}

func uniqueUsername() string {
	return "user-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:20]
}

func mustUsername(t *testing.T, raw string) domain.Username {
	t.Helper()
	username, err := domain.NewUsername(raw)
	require.NoError(t, err)
	return username
}

func addCredentials(t *testing.T, raw, hash string) *domain.Credentials {
	t.Helper()

	credentials, err := domain.LoadCredentials(uuid.New(), mustUsername(t, raw), hash)
	require.NoError(t, err)
	require.NoError(t, credentialsRepository(t).Add(credentials))

	return credentials
}

func requireSameCredentials(t *testing.T, want, got *domain.Credentials) {
	t.Helper()

	require.Equal(t, want.UserID(), got.UserID())
	require.Equal(t, want.Username(), got.Username())
	require.Equal(t, want.PasswordHash(), got.PasswordHash())
}

func TestCredentialsRepositoryAddAndFind(t *testing.T) {
	credentials := addCredentials(t, uniqueUsername(), "hash-1")

	byUsername, err := credentialsRepository(t).FindByUsername(credentials.Username())
	require.NoError(t, err)
	requireSameCredentials(t, credentials, byUsername)

	byUserID, err := credentialsRepository(t).FindByUserIDForUpdate(credentials.UserID())
	require.NoError(t, err)
	requireSameCredentials(t, credentials, byUserID)
}

func TestCredentialsRepositoryNotFound(t *testing.T) {
	_, err := credentialsRepository(t).FindByUsername(mustUsername(t, uniqueUsername()))
	require.ErrorIs(t, err, domain.ErrCredentialsNotFound)

	_, err = credentialsRepository(t).FindByUserIDForUpdate(uuid.New())
	require.ErrorIs(t, err, domain.ErrCredentialsNotFound)
}

func TestCredentialsRepositoryUsernameTaken(t *testing.T) {
	existing := addCredentials(t, uniqueUsername(), "hash-1")

	duplicate, err := domain.LoadCredentials(uuid.New(), existing.Username(), "hash-2")
	require.NoError(t, err)

	require.ErrorIs(t, credentialsRepository(t).Add(duplicate), domain.ErrUsernameTaken)
}

func TestCredentialsRepositoryUpdate(t *testing.T) {
	credentials := addCredentials(t, uniqueUsername(), "old-hash")

	changed, err := domain.LoadCredentials(credentials.UserID(), credentials.Username(), "new-hash")
	require.NoError(t, err)
	require.NoError(t, credentialsRepository(t).Update(changed))

	stored, err := credentialsRepository(t).FindByUsername(credentials.Username())
	require.NoError(t, err)
	require.Equal(t, "new-hash", stored.PasswordHash())
}
