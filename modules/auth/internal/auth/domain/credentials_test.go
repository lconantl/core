package domain_test

import (
	"auth/internal/auth/domain"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

const (
	testUsername = "antllcon"
	oldPlain     = "old-password"
	newPlain     = "new-password"
	wrongPlain   = "wrong-password"

	hashPrefix = "hashed:"
	oldHash    = hashPrefix + oldPlain
	newHash    = hashPrefix + newPlain
)

var errHasherFailed = errors.New("хешер сломался")

// fakeHasher: быстрый и предсказуемый хешер для доменных тестов.
type fakeHasher struct {
	hashErr error
}

func (hasher fakeHasher) Hash(password domain.Password) (string, error) {
	if hasher.hashErr != nil {
		return "", hasher.hashErr
	}
	return hashPrefix + password.Plain(), nil
}

func (hasher fakeHasher) Verify(plain, hash string) error {
	if hash != hashPrefix+plain {
		return domain.ErrWrongPassword
	}
	return nil
}

func mustUsername(test *testing.T, raw string) domain.Username {
	test.Helper()

	username, err := domain.NewUsername(raw)
	require.NoError(test, err)

	return username
}

func mustPassword(test *testing.T, raw string) domain.Password {
	test.Helper()

	password, err := domain.NewPassword(raw)
	require.NoError(test, err)

	return password
}

func newCredentials(test *testing.T) *domain.Credentials {
	test.Helper()

	credentials, err := domain.NewCredentials(uuid.New(), mustUsername(test, testUsername), mustPassword(test, oldPlain), fakeHasher{})
	require.NoError(test, err)

	return credentials
}

func TestNewCredentials(test *testing.T) {
	id := uuid.New()

	credentials, err := domain.NewCredentials(id, mustUsername(test, testUsername), mustPassword(test, oldPlain), fakeHasher{})

	require.NoError(test, err)
	require.Equal(test, id, credentials.UserID())
	require.Equal(test, testUsername, credentials.Username().String())
	require.Equal(test, oldHash, credentials.PasswordHash(), "в Credentials хранится хеш, а не пароль")
}

type newCredentialsCase struct {
	name     string
	id       uuid.UUID
	username domain.Username
	password domain.Password
	hasher   domain.PasswordHasher
	err      error
}

func checkNewCredentials(test *testing.T, testCase newCredentialsCase) {
	test.Helper()

	credentials, err := domain.NewCredentials(testCase.id, testCase.username, testCase.password, testCase.hasher)

	require.ErrorIs(test, err, testCase.err)
	require.Nil(test, credentials)
}

func TestNewCredentialsErrors(test *testing.T) {
	username := mustUsername(test, testUsername)
	password := mustPassword(test, oldPlain)

	testCases := []newCredentialsCase{
		{"пустой ID", uuid.Nil(), username, password, fakeHasher{}, domain.ErrInvalidUserID},
		{"пустой логин", uuid.New(), domain.Username{}, password, fakeHasher{}, domain.ErrInvalidUsername},
		{"пустой пароль", uuid.New(), username, domain.Password{}, fakeHasher{}, domain.ErrInvalidPassword},
		{"хешер вернул ошибку", uuid.New(), username, password, fakeHasher{hashErr: errHasherFailed}, errHasherFailed},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkNewCredentials(test, testCase)
		})
	}
}

type loadCredentialsCase struct {
	name     string
	id       uuid.UUID
	username domain.Username
	hash     string
	err      error
}

func checkLoadCredentials(test *testing.T, testCase loadCredentialsCase) {
	test.Helper()

	credentials, err := domain.LoadCredentials(testCase.id, testCase.username, testCase.hash)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
		require.Nil(test, credentials)
		return
	}

	require.NoError(test, err)
	require.Equal(test, testCase.id, credentials.UserID())
	require.Equal(test, testCase.username, credentials.Username())
	require.Equal(test, testCase.hash, credentials.PasswordHash())
}

func TestLoadCredentials(test *testing.T) {
	id := uuid.New()
	username := mustUsername(test, testUsername)

	testCases := []loadCredentialsCase{
		{"загрузка из хранилища", id, username, oldHash, nil},
		{"пустой ID", uuid.Nil(), username, oldHash, domain.ErrInvalidUserID},
		{"пустой логин", id, domain.Username{}, oldHash, domain.ErrInvalidUsername},
		{"пустой хеш", id, username, "", domain.ErrInvalidPasswordHash},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkLoadCredentials(test, testCase)
		})
	}
}

func TestVerifyPassword(test *testing.T) {
	credentials := newCredentials(test)

	require.NoError(test, credentials.VerifyPassword(oldPlain, fakeHasher{}))
	require.ErrorIs(test, credentials.VerifyPassword(wrongPlain, fakeHasher{}), domain.ErrWrongPassword)
	require.ErrorIs(test, credentials.VerifyPassword("", fakeHasher{}), domain.ErrWrongPassword)
}

type changePasswordCase struct {
	name     string
	old      string
	new      domain.Password
	hasher   domain.PasswordHasher
	wantHash string
	err      error
}

func checkChangePassword(test *testing.T, testCase changePasswordCase) {
	test.Helper()

	credentials := newCredentials(test)

	err := credentials.ChangePassword(testCase.old, testCase.new, testCase.hasher)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
	} else {
		require.NoError(test, err)
	}
	require.Equal(test, testCase.wantHash, credentials.PasswordHash())
}

func TestChangePassword(test *testing.T) {
	newPassword := mustPassword(test, newPlain)

	testCases := []changePasswordCase{
		{"смена пароля", oldPlain, newPassword, fakeHasher{}, newHash, nil},
		{"неверный старый пароль", wrongPlain, newPassword, fakeHasher{}, oldHash, domain.ErrWrongPassword},
		{"пустой новый пароль", oldPlain, domain.Password{}, fakeHasher{}, oldHash, domain.ErrInvalidPassword},
		{"хешер сломался", oldPlain, newPassword, fakeHasher{hashErr: errHasherFailed}, oldHash, errHasherFailed},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkChangePassword(test, testCase)
		})
	}
}
