package password_test

import (
	"testing"

	"auth/internal/auth/infra/password"
)

func TestHashAndVerify(test *testing.T) {
	hash, err := password.Hash("my-password")

	if err != nil {
		test.Fatal(err)
	}

	if hash == "my-password" {
		test.Fatal("Пароль не должен храниться в открытом виде")
	}

	err = password.Verify("my-password", hash)

	if err != nil {
		test.Fatal("Правильный пароль должен пройти проверку")
	}
}

func TestVerifyWrongPassword(test *testing.T) {
	hash, err := password.Hash("my-password")

	if err != nil {
		test.Fatal(err)
	}

	err = password.Verify("wrong-password", hash)

	if err == nil {
		test.Fatal("Неправильный пароль не должен пройти проверку")
	}
}

func TestHashEmptyPassword(test *testing.T) {
	hash, err := password.Hash("")

	if err != password.ErrEmptyPassword {
		test.Fatal("Ожидалась ошибка пустого пароля")
	}

	if hash != "" {
		test.Fatal("Хэш не должен создаваться")
	}
}
