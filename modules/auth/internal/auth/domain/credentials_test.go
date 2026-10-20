package domain_test

import (
	"auth/internal/auth/domain"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

var errHasherFailed = errors.New("хешер сломался")

type fakeHasher struct {
	hashErr error
}

func (hasher fakeHasher) Hash(password domain.Password) (string, error) {
	if hasher.hashErr != nil {
		return "", hasher.hashErr
	}
	return "hashed:" + password.Plain(), nil
}

func (hasher fakeHasher) Verify(plain string, hash string) error {
	if hash != "hashed:"+plain {
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

	credentials, err := domain.NewCredentials(uuid.New(), mustUsername(test, "antllcon"), mustPassword(test, "old-password"), fakeHasher{})
	require.NoError(test, err)

	return credentials
}

func TestNewCredentials(test *testing.T) {
	id := uuid.New()

	credentials, err := domain.NewCredentials(id, mustUsername(test, "antllcon"), mustPassword(test, "old-password"), fakeHasher{})

	require.NoError(test, err)
	require.Equal(test, id, credentials.UserID())
	require.Equal(test, "antllcon", credentials.Username().String())
	require.Equal(test, "hashed:old-password", credentials.PasswordHash(), "в Credentials хранится хеш, а не пароль")
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
	username := mustUsername(test, "antllcon")
	password := mustPassword(test, "old-password")

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
	username := mustUsername(test, "antllcon")

	testCases := []loadCredentialsCase{
		{"загрузка из хранилища", id, username, "hashed:old-password", nil},
		{"пустой ID", uuid.Nil(), username, "hashed:old-password", domain.ErrInvalidUserID},
		{"пустой логин", id, domain.Username{}, "hashed:old-password", domain.ErrInvalidUsername},
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

	require.NoError(test, credentials.VerifyPassword("old-password", fakeHasher{}))
	require.ErrorIs(test, credentials.VerifyPassword("wrong-password", fakeHasher{}), domain.ErrWrongPassword)
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
	newPassword := mustPassword(test, "new-password")

	testCases := []changePasswordCase{
		{"смена пароля", "old-password", newPassword, fakeHasher{}, "hashed:new-password", nil},
		{"неверный старый пароль", "wrong-password", newPassword, fakeHasher{}, "hashed:old-password", domain.ErrWrongPassword},
		{"пустой новый пароль", "old-password", domain.Password{}, fakeHasher{}, "hashed:old-password", domain.ErrInvalidPassword},
		{"хешер сломался", "old-password", newPassword, fakeHasher{hashErr: errHasherFailed}, "hashed:old-password", errHasherFailed},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkChangePassword(test, testCase)
		})
	}
}
