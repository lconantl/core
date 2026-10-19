package domain_test

import (
	"identity/internal/identity/domain"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

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

	user, err := domain.LoadUser(uuid.New(), testCase.from, domain.Profile{})
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
