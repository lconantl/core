package domain_test

import (
	"auth/internal/auth/domain"
	"testing"
	"uuid"
)

func TestNewCredentials(test *testing.T) {
	id := uuid.New()
	username := "antllcon"
	hash := "test-hash"

	credentials, err := domain.NewCredentials(id, username, hash)

	if err != nil {
		test.Fatal(err)
	}

	if credentials.UserID() != id {
		test.Fatal("ID пользователя не совпадает")
	}

	if credentials.Username() != username {
		test.Fatal("Username пользователя не совпадает")
	}

	if credentials.PasswordHash() != hash {
		test.Fatal("Хэш пароля не совпадает")
	}
}

func TestNewCredentialsInvalidID(test *testing.T) {
	credentials, err := domain.NewCredentials(
		uuid.Nil(),
		"antllcon",
		"test-hash",
	)

	if err != domain.ErrInvalidUserID {
		test.Fatal("Ожидалась ошибка некорректного ID")
	}

	if credentials != nil {
		test.Fatal("Credentials не должны создаваться")
	}
}

func TestNewCredentialsSpacesUsername(test *testing.T) {
	credentials, err := domain.NewCredentials(
		uuid.New(),
		"   ",
		"test-hash",
	)

	if err != domain.ErrInvalidUsername {
		test.Fatal("Username из пробелов должен быть запрещён")
	}

	if credentials != nil {
		test.Fatal("Credentials не должны создаваться")
	}
}

func TestNewCredentialsEmptyPasswordHash(test *testing.T) {
	credentials, err := domain.NewCredentials(
		uuid.New(),
		"antllcon",
		"",
	)

	if err != domain.ErrInvalidPasswordHash {
		test.Fatal("Пустой хэш пароля должен быть запрещён")
	}

	if credentials != nil {
		test.Fatal("Credentials не должны создаваться")
	}
}
