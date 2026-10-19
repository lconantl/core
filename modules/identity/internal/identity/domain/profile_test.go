package domain_test

import (
	"identity/internal/identity/domain"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func validFields() domain.ProfileFields {
	return domain.ProfileFields{
		FirstName: "Степан",
		LastName:  "Иванов",
		Email:     "antllcon.core@email.com",
	}
}

func validProfile(test *testing.T) domain.Profile {
	test.Helper()

	profile, err := domain.NewProfile(validFields())
	require.NoError(test, err)

	return profile
}

func with(change func(fields *domain.ProfileFields)) domain.ProfileFields {
	fields := validFields()
	change(&fields)
	return fields
}

type profileCase struct {
	name   string
	fields domain.ProfileFields
	errs   []error
}

func checkProfile(test *testing.T, testCase profileCase) {
	test.Helper()

	profile, err := domain.NewProfile(testCase.fields)

	if len(testCase.errs) == 0 {
		require.NoError(test, err)
		require.False(test, profile.IsZero())
		return
	}

	for _, want := range testCase.errs {
		require.ErrorIs(test, err, want)
	}
	require.True(test, profile.IsZero())
}

func TestNewProfile(test *testing.T) {
	testCases := []profileCase{
		{"минимальный профиль", validFields(), nil},

		{"пустое имя", with(func(f *domain.ProfileFields) { f.FirstName = "" }), []error{domain.ErrInvalidFirstName}},
		{"имя из пробелов", with(func(f *domain.ProfileFields) { f.FirstName = "   " }), []error{domain.ErrInvalidFirstName}},
		{"имя ровно 64 символа", with(func(f *domain.ProfileFields) { f.FirstName = strings.Repeat("я", 64) }), nil},
		{"имя 65 символов", with(func(f *domain.ProfileFields) { f.FirstName = strings.Repeat("я", 65) }), []error{domain.ErrInvalidFirstName}},
		{"пустая фамилия", with(func(f *domain.ProfileFields) { f.LastName = "" }), []error{domain.ErrInvalidLastName}},

		{"пустая почта", with(func(f *domain.ProfileFields) { f.Email = "" }), []error{domain.ErrInvalidEmail}},
		{"почта без @", with(func(f *domain.ProfileFields) { f.Email = "antllcon.email.com" }), []error{domain.ErrInvalidEmail}},
		{"почта с именем", with(func(f *domain.ProfileFields) { f.Email = "Степан <a@b.ru>" }), []error{domain.ErrInvalidEmail}},
		{"почта длиннее 254", with(func(f *domain.ProfileFields) { f.Email = strings.Repeat("a", 250) + "@b.ru" }), []error{domain.ErrInvalidEmail}},

		{"телефон пустой", with(func(f *domain.ProfileFields) { f.Phone = "" }), nil},
		{"телефон без плюса", with(func(f *domain.ProfileFields) { f.Phone = "89123456789" }), []error{domain.ErrInvalidPhone}},
		{"телефон слишком короткий", with(func(f *domain.ProfileFields) { f.Phone = "+12" }), []error{domain.ErrInvalidPhone}},
		{"телефон с буквами", with(func(f *domain.ProfileFields) { f.Phone = "+7abc" }), []error{domain.ErrInvalidPhone}},

		{"описание ровно 70 символов", with(func(f *domain.ProfileFields) { f.Description = strings.Repeat("я", 70) }), nil},
		{"описание 71 символ", with(func(f *domain.ProfileFields) { f.Description = strings.Repeat("я", 71) }), []error{domain.ErrInvalidDescription}},

		{"язык пустой", with(func(f *domain.ProfileFields) { f.LanguageCode = "" }), nil},
		{"язык не поддерживается", with(func(f *domain.ProfileFields) { f.LanguageCode = "de" }), []error{domain.ErrInvalidLanguageCode}},

		{"несколько ошибок сразу", domain.ProfileFields{}, []error{domain.ErrInvalidFirstName, domain.ErrInvalidLastName, domain.ErrInvalidEmail}},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkProfile(test, testCase)
		})
	}
}

func TestNewProfileNormalization(test *testing.T) {
	avatarID := uuid.New()

	profile, err := domain.NewProfile(domain.ProfileFields{
		FirstName:    "  Степан  ",
		LastName:     "	 Смирнов ",
		Email:        " 	 Antllcon.Core@Email.COM ",
		Phone:        "+7 912 345-67-89",
		Description:  "  Какой-то текст  ",
		LanguageCode: " RU ",
		AvatarID:     avatarID,
	})
	require.NoError(test, err)

	require.Equal(test, "Степан", profile.FirstName())
	require.Equal(test, "Смирнов", profile.LastName())
	require.Equal(test, "antllcon.core@email.com", profile.Email())
	require.Equal(test, "+79123456789", profile.Phone())
	require.Equal(test, "Какой-то текст", profile.Description())
	require.Equal(test, "ru", profile.LanguageCode())
	require.Equal(test, avatarID, profile.AvatarID())
	require.True(test, profile.HasAvatar())
}

func TestProfileWithoutAvatar(test *testing.T) {
	require.False(test, validProfile(test).HasAvatar())
}
