package domain_test

import (
	"identity/internal/identity/domain"
	"testing"
	"uuid"
)

func TestNewUser(test *testing.T) {
	id := uuid.New()
	profile := domain.Profile{
		Username: "antllcon",
		Email:    "antllcon.core@email.com",
	}

	user, err := domain.NewUser(id, profile)

	if err != nil {
		test.Fatal(err)
	}

	if user.ID() != id {
		test.Fatal("ID пользователя не совпадает")
	}

	if user.Status() != domain.UserStatusPending {
		test.Fatal("Новый пользователь должен быть Pending")
	}

	if user.Profile() != profile {
		test.Fatal("Профиль пользователя не совпадает")
	}
}

func TestNewUserInvalidID(test *testing.T) {
	user, err := domain.NewUser(uuid.Nil(), domain.Profile{})

	if err != domain.ErrorInvalidUserID {
		test.Fatal("Ожидалась ошибка некорректного ID")
	}

	if user != nil {
		test.Fatal("Пользователь не должен создаваться")
	}
}

func TestRestoreUser(test *testing.T) {
	id := uuid.New()
	profile := domain.Profile{Username: "antllcon"}

	user, err := domain.RestoreUser(id, domain.UserStatusActive, profile)

	if err != nil {
		test.Fatal(err)
	}

	if user.ID() != id {
		test.Fatal("ID пользователя изменился")
	}

	if user.Status() != domain.UserStatusActive {
		test.Fatal("Статус пользователя не восстановился")
	}

	if user.Profile() != profile {
		test.Fatal("Профиль пользователя не восстановился")
	}
}

func TestRestoreUserInvalidID(test *testing.T) {
	user, err := domain.RestoreUser(
		uuid.Nil(),
		domain.UserStatusActive,
		domain.Profile{},
	)

	if err != domain.ErrorInvalidUserID {
		test.Fatal("Ожидалась ошибка некорректного ID")
	}

	if user != nil {
		test.Fatal("Пользователь не должен восстанавливаться")
	}
}

func TestRestoreUserInvalidStatus(test *testing.T) {
	user, err := domain.RestoreUser(
		uuid.New(),
		"deleted",
		domain.Profile{},
	)

	if err != domain.ErrorInvalidUserStatus {
		test.Fatal("Ожидалась ошибка некорректного статуса")
	}

	if user != nil {
		test.Fatal("Пользователь не должен восстанавливаться")
	}

	user, err = domain.RestoreUser(
		uuid.New(),
		"",
		domain.Profile{},
	)

	if err != domain.ErrorInvalidUserStatus || user != nil {
		test.Fatal("Пустой статус должен быть запрещён")
	}
}

func TestUserActivation(test *testing.T) {
	user, err := domain.NewUser(uuid.New(), domain.Profile{})
	if err != nil {
		test.Fatal(err)
	}

	if user.Status() != domain.UserStatusPending {
		test.Fatal("Начальный статус должен быть Pending")
	}

	err = user.Activate()
	if err != nil {
		test.Fatal(err)
	}

	if user.Status() != domain.UserStatusActive {
		test.Fatal("После активации статус должен быть Active")
	}

	err = user.Activate()
	if err != nil {
		test.Fatal("Повторная активация должна быть успешной")
	}

	if user.Status() != domain.UserStatusActive {
		test.Fatal("Повторная активация не должна менять статус")
	}
}

func TestUpdateProfile(test *testing.T) {
	profile := domain.Profile{
		Username: "antllcon",
		Email:    "antllcon.core@email.com",
	}

	user, err := domain.NewUser(uuid.New(), profile)
	if err != nil {
		test.Fatal(err)
	}

	id := user.ID()
	status := user.Status()

	updated := profile
	updated.AvatarURL = "avatar.png"

	user.UpdateProfile(updated)

	if user.Profile() != updated {
		test.Fatal("Профиль не обновился")
	}

	if user.ID() != id {
		test.Fatal("Обновление профиля изменило ID")
	}

	if user.Status() != status {
		test.Fatal("Обновление профиля изменило статус")
	}

	user.UpdateProfile(domain.Profile{})

	if user.Profile() != (domain.Profile{}) {
		test.Fatal("Профиль должен очищаться")
	}
}

func TestProfileReturnsCopy(test *testing.T) {
	const username = "antllcon"

	profile := domain.Profile{
		Username: username,
	}

	user, err := domain.NewUser(uuid.New(), profile)
	if err != nil {
		test.Fatal(err)
	}

	copyProfile := user.Profile()
	copyProfile.Username = "another"

	if user.Profile().Username != username {
		test.Fatal("Изменение копии затронуло профиль пользователя")
	}
}
