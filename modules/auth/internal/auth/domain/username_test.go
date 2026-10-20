package domain_test

import (
	"auth/internal/auth/domain"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type usernameCase struct {
	name string
	raw  string
	want string
	err  error
}

func checkUsername(test *testing.T, testCase usernameCase) {
	test.Helper()

	username, err := domain.NewUsername(testCase.raw)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
		require.True(test, username.IsZero())
		return
	}

	require.NoError(test, err)
	require.Equal(test, testCase.want, username.String())
}

func TestNewUsername(test *testing.T) {
	testCases := []usernameCase{
		{"обычный логин", "antllcon", "antllcon", nil},
		{"с цифрами и дефисом", "ant-llcon-1", "ant-llcon-1", nil},
		{"обрезаются пробелы и регистр", "  AntLLcon  ", "antllcon", nil},
		{"ровно 3 символа", "abc", "abc", nil},
		{"ровно 32 символа", "a" + strings.Repeat("b", 31), "a" + strings.Repeat("b", 31), nil},

		{"пустой", "", "", domain.ErrInvalidUsername},
		{"только пробелы", "   ", "", domain.ErrInvalidUsername},
		{"2 символа", "ab", "", domain.ErrInvalidUsername},
		{"33 символа", "a" + strings.Repeat("b", 32), "", domain.ErrInvalidUsername},
		{"начинается с цифры", "1abc", "", domain.ErrInvalidUsername},
		{"начинается с дефиса", "-abc", "", domain.ErrInvalidUsername},
		{"подчёркивание", "ant_llcon", "", domain.ErrInvalidUsername},
		{"пробел внутри", "ant llcon", "", domain.ErrInvalidUsername},
		{"кириллица", "антон", "", domain.ErrInvalidUsername},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkUsername(test, testCase)
		})
	}
}
