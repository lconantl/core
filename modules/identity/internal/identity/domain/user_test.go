package domain_test

import (
	"identity/internal/identity/domain"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func validProfile(test *testing.T) domain.Profile {
	test.Helper()

	profile, err := domain.NewProfile(domain.ProfileFields{
		FirstName: "Иван",
		LastName:  "Смирнов",
		Email:     "antllcon.core@email.com",
	})
	require.NoError(test, err)

	return profile
}

type transition struct {
	name   string
	from   domain.UserStatus
	action func(*domain.User) error
	to     domain.UserStatus
	err    error
}

func activate(user *domain.User) error {
	return user.Activate()
}
func remove(user *domain.User) error {
	user.Delete()
	return nil
}

func check(test *testing.T, testCase transition) {
	test.Helper()

	user, err := domain.LoadUser(uuid.New(), testCase.from, validProfile(test))
	require.NoError(test, err)

	err = testCase.action(user)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
	} else {
		require.NoError(test, err)
	}

	require.Equal(test, testCase.to, user.Status())
}

func TestUserStatus(test *testing.T) {
	testCases := []transition{
		{"закончил регистрацию аккаунта", domain.UserStatusPending, activate, domain.UserStatusActive, nil},
		{"забраковал регистрацию аккаунта", domain.UserStatusPending, remove, domain.UserStatusDeleted, nil},
		{"повторная активация аккаунта", domain.UserStatusActive, activate, domain.UserStatusActive, nil},
		{"удаление аккаунта", domain.UserStatusActive, remove, domain.UserStatusDeleted, nil},
		{"активация удаленного аккаунта", domain.UserStatusDeleted, activate, domain.UserStatusDeleted, domain.ErrInvalidStatusTransition},
		{"удаление удаленного аккаунта", domain.UserStatusDeleted, remove, domain.UserStatusDeleted, nil},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			check(test, testCase)
		})
	}
}

type load struct {
	name    string
	id      uuid.UUID
	status  domain.UserStatus
	profile domain.Profile
	err     error
}

func checkLoad(test *testing.T, testCase load) {
	test.Helper()

	user, err := domain.LoadUser(testCase.id, testCase.status, testCase.profile)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
		require.Nil(test, user)
		return
	}

	require.NoError(test, err)
	require.Equal(test, testCase.id, user.ID())
	require.Equal(test, testCase.status, user.Status())
	require.Equal(test, testCase.profile, user.Profile())
}

func TestLoadUser(test *testing.T) {
	id := uuid.New()
	profile := validProfile(test)

	testCases := []load{
		{"загрузка ожидающего аккаунта", id, domain.UserStatusPending, profile, nil},
		{"загрузка активного аккаунта", id, domain.UserStatusActive, profile, nil},
		{"загрузка удаленного аккаунта", id, domain.UserStatusDeleted, profile, nil},
		{"пустой статус", id, "", profile, domain.ErrInvalidUserStatus},
		{"неизвестный статус", id, "blocked", profile, domain.ErrInvalidUserStatus},
		{"статус в другом регистре", id, "Active", profile, domain.ErrInvalidUserStatus},
		{"пустой ID", uuid.Nil(), domain.UserStatusActive, profile, domain.ErrInvalidUserID},
		{"пустой профиль", id, domain.UserStatusActive, domain.Profile{}, domain.ErrInvalidProfile},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkLoad(test, testCase)
		})
	}
}

func TestNewUser(test *testing.T) {
	user, err := domain.NewUser(uuid.New(), domain.Profile{})

	require.NoError(test, err)
	require.Equal(test, domain.UserStatusPending, user.Status())
}
