package domain_test

import (
	"auth/internal/auth/domain"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type passwordCase struct {
	name string
	raw  string
	err  error
}

func checkPassword(test *testing.T, testCase passwordCase) {
	test.Helper()

	password, err := domain.NewPassword(testCase.raw)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
		require.True(test, password.IsZero())
		return
	}

	require.NoError(test, err)
	require.Equal(test, testCase.raw, password.Plain(), "пароль хранится как есть, без trim")
}

func TestNewPassword(test *testing.T) {
	testCases := []passwordCase{
		{"ровно 8 символов", "12345678", nil},
		{"8 символов кириллицей (14 байт)", "пароль12", nil},
		{"пробелы сохраняются", "  pass word  ", nil},
		{"ровно 72 байта", strings.Repeat("a", 72), nil},

		{"пустой", "", domain.ErrPasswordTooShort},
		{"7 символов", "1234567", domain.ErrPasswordTooShort},
		{"73 байта", strings.Repeat("a", 73), domain.ErrPasswordTooLong},
		{"37 символов кириллицей (74 байта)", strings.Repeat("я", 37), domain.ErrPasswordTooLong},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkPassword(test, testCase)
		})
	}
}

func TestPasswordHiddenInLogs(test *testing.T) {
	password, err := domain.NewPassword("super-secret")
	require.NoError(test, err)

	require.NotContains(test, fmt.Sprint(password), "super-secret")
	require.NotContains(test, fmt.Sprintf("%v %s", password, password), "super-secret")
}
