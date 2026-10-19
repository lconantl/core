package domain

import (
	"errors"
	"net/mail"
	"slices"
	"strings"
	"unicode/utf8"
	"uuid"

	"github.com/nyaruka/phonenumbers"
)

const (
	maxNameLength        = 64
	maxEmailLength       = 254
	maxDescriptionLength = 70
)

var supportedLanguages = []string{"ru", "en"}

var (
	ErrInvalidProfile      = errors.New("профиль не заполнен")
	ErrInvalidFirstName    = errors.New("некорректное имя")
	ErrInvalidLastName     = errors.New("некорректная фамилия")
	ErrInvalidEmail        = errors.New("некорректная почта")
	ErrInvalidPhone        = errors.New("некорректный телефон")
	ErrInvalidDescription  = errors.New("слишком длинное описание")
	ErrInvalidLanguageCode = errors.New("неподдерживаемый язык")
)

type ProfileFields struct {
	FirstName    string
	LastName     string
	Email        string
	Phone        string
	Description  string
	LanguageCode string
	AvatarID     uuid.UUID
}

type Profile struct {
	firstName    string
	lastName     string
	phone        string
	email        string
	description  string
	languageCode string
	avatarID     uuid.UUID
}

func NewProfile(fields ProfileFields) (Profile, error) {
	var errs []error

	collect := func(value string, err error) string {
		if err != nil {
			errs = append(errs, err)
		}

		return value
	}

	profile := Profile{
		firstName:    collect(normalizeName(fields.FirstName, ErrInvalidFirstName)),
		lastName:     collect(normalizeName(fields.LastName, ErrInvalidLastName)),
		email:        collect(normalizeEmail(fields.Email)),
		phone:        collect(normalizePhone(fields.Phone)),
		description:  collect(normalizeDescription(fields.Description)),
		languageCode: collect(normalizeLanguageCode(fields.LanguageCode)),
		avatarID:     fields.AvatarID,
	}

	if err := errors.Join(errs...); err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func (profile Profile) FirstName() string {
	return profile.firstName
}

func (profile Profile) LastName() string {
	return profile.lastName
}

func (profile Profile) Email() string {
	return profile.email
}

func (profile Profile) Phone() string {
	return profile.phone
}

func (profile Profile) Description() string {
	return profile.description
}

func (profile Profile) LanguageCode() string {
	return profile.languageCode
}

func (profile Profile) AvatarID() uuid.UUID {
	return profile.avatarID
}

func (profile Profile) HasAvatar() bool {
	return profile.avatarID != uuid.Nil()
}

func (profile Profile) IsZero() bool {
	return profile == Profile{}
}

func normalizeName(value string, errInvalid error) (string, error) {
	value = strings.TrimSpace(value)
	length := utf8.RuneCountInString(value)

	if length == 0 || length > maxNameLength {
		return "", errInvalid
	}

	return value, nil
}

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	if value == "" || len(value) > maxEmailLength {
		return "", ErrInvalidEmail
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", ErrInvalidEmail
	}

	return value, nil
}

func normalizePhone(value string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", nil
	}

	number, err := phonenumbers.Parse(value, "")
	if err != nil || !phonenumbers.IsValidNumber(number) {
		return "", ErrInvalidPhone
	}

	return phonenumbers.Format(number, phonenumbers.E164), nil
}

func normalizeDescription(value string) (string, error) {
	value = strings.TrimSpace(value)

	if utf8.RuneCountInString(value) > maxDescriptionLength {
		return "", ErrInvalidDescription
	}

	return value, nil
}

func normalizeLanguageCode(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	if value == "" {
		return "", nil
	}

	if !slices.Contains(supportedLanguages, value) {
		return "", ErrInvalidLanguageCode
	}

	return value, nil
}
