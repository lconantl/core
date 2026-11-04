package integration_tests

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"identity/internal/identity/domain"
	"identity/internal/identity/infra/mysql/repository"
)

func userRepository(t *testing.T) domain.UserRepository {
	t.Helper()
	return repository.NewUserRepository(t.Context(), testEnv.connector.TransactionalClient())
}

func uniqueEmail() string {
	return "user-" + uuid.New().String() + "@example.com"
}

func newProfile(t *testing.T, email string) domain.Profile {
	t.Helper()

	profile, err := domain.NewProfile(domain.ProfileFields{
		FirstName: "Степан",
		LastName:  "Иванов",
		Email:     email,
	})
	require.NoError(t, err)

	return profile
}

func addUser(t *testing.T, email string) *domain.User {
	t.Helper()

	user, err := domain.NewUser(uuid.New(), newProfile(t, email))
	require.NoError(t, err)
	require.NoError(t, userRepository(t).Add(user))

	return user
}

func TestUserRepositoryAddAndFind(t *testing.T) {
	profile, err := domain.NewProfile(domain.ProfileFields{
		FirstName:    "Степан",
		LastName:     "Иванов",
		Email:        uniqueEmail(),
		Phone:        "+79123456789",
		Description:  "Мяу мяу мяу, спать хочу",
		LanguageCode: "ru",
		AvatarID:     uuid.New(),
	})
	require.NoError(t, err)

	user, err := domain.NewUser(uuid.New(), profile)
	require.NoError(t, err)
	require.NoError(t, userRepository(t).Add(user))

	stored, err := userRepository(t).FindByID(user.ID())
	require.NoError(t, err)
	require.Equal(t, user.ID(), stored.ID())
	require.Equal(t, domain.UserStatusPending, stored.Status())
	require.Equal(t, profile, stored.Profile())
}

func TestUserRepositoryWithoutOptionalFields(t *testing.T) {
	user := addUser(t, uniqueEmail())

	stored, err := userRepository(t).FindByIDForUpdate(user.ID())
	require.NoError(t, err)
	require.Equal(t, user.Profile(), stored.Profile())
	require.False(t, stored.Profile().HasAvatar())
}

func TestUserRepositoryNotFound(t *testing.T) {
	_, err := userRepository(t).FindByID(uuid.New())
	require.ErrorIs(t, err, domain.ErrUserNotFound)

	_, err = userRepository(t).FindByIDForUpdate(uuid.New())
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepositoryUpdate(t *testing.T) {
	user := addUser(t, uniqueEmail())

	require.NoError(t, user.Activate())
	newEmail := uniqueEmail()
	require.NoError(t, user.UpdateProfile(newProfile(t, newEmail)))
	require.NoError(t, userRepository(t).Update(user))

	stored, err := userRepository(t).FindByID(user.ID())
	require.NoError(t, err)
	require.Equal(t, domain.UserStatusActive, stored.Status())
	require.Equal(t, newEmail, stored.Profile().Email())
}

func TestUserRepositoryEmailTaken(t *testing.T) {
	email := uniqueEmail()
	addUser(t, email)

	duplicate, err := domain.NewUser(uuid.New(), newProfile(t, email))
	require.NoError(t, err)
	require.ErrorIs(t, userRepository(t).Add(duplicate), domain.ErrEmailTaken)

	other := addUser(t, uniqueEmail())
	require.NoError(t, other.UpdateProfile(newProfile(t, email)))
	require.ErrorIs(t, userRepository(t).Update(other), domain.ErrEmailTaken)
}

func TestUserRepositoryDeletedKeepsEmail(t *testing.T) {
	email := uniqueEmail()
	deleted := addUser(t, email)

	deleted.Delete()
	require.NoError(t, userRepository(t).Update(deleted))

	stored, err := userRepository(t).FindByID(deleted.ID())
	require.NoError(t, err)
	require.Equal(t, domain.UserStatusDeleted, stored.Status())

	duplicate, err := domain.NewUser(uuid.New(), newProfile(t, email))
	require.NoError(t, err)
	require.ErrorIs(t, userRepository(t).Add(duplicate), domain.ErrEmailTaken)
}
