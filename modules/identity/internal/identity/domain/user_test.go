package domain_test

import (
	"identity/internal/identity/domain"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func activate(user *domain.User) error {
	return user.Activate()
}
func remove(user *domain.User) error {
	user.Delete()
	return nil
}

func TestUserStatusTransitions(test *testing.T) {
	tests := []struct {
		name   string
		from   domain.UserStatus
		action func(*domain.User) error
		to     domain.UserStatus
		err    error
	}{
		{"закончил регистрацию аккаунта", domain.UserStatusPending, activate, domain.UserStatusActive, nil},
		{"забраковал регистрацию аккаунта", domain.UserStatusPending, remove, domain.UserStatusDeleted, nil},
		{"повторная активация аккаунта (случайно)", domain.UserStatusActive, activate, domain.UserStatusActive, nil},
		{"удаление аккаунта", domain.UserStatusActive, remove, domain.UserStatusDeleted, domain.ErrInvalidUserStatus},
		{"активация удаленного аккаунта", domain.UserStatusDeleted, activate, domain.UserStatusDeleted, domain.ErrInvalidStatusTransition},
		{"удаление удаленного аккаунта", domain.UserStatusDeleted, remove, domain.UserStatusDeleted, nil},
	}

	for _, i := range tests {
		test.Run(i.name, func(t *testing.T) {
			user, err := domain.LoadUser(uuid.New(), i.from, domain.Profile{})
			require.NoError(t, err)

			err = i.action(user)

			if i.err != nil {
				require.ErrorIs(t, err, i.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, i.to, user.Status())
		})
	}
}
