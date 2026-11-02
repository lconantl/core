package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"
)

const maxDeviceLength = 255

var (
	ErrInvalidSessionID    = errors.New("некорректный ID сессии")
	ErrInvalidUserID       = errors.New("некорректный ID пользователя")
	ErrInvalidTokenHash    = errors.New("некорректный хеш токена")
	ErrInvalidTTL          = errors.New("некорректный срок жизни")
	ErrInvalidSessionTime  = errors.New("некорректное время сессии")
	ErrSessionRevoked      = errors.New("сессия отозвана")
	ErrSessionExpired      = errors.New("сессия истекла")
	ErrInvalidRefreshToken = errors.New("неверный refresh-токен")
	ErrRefreshTokenReused  = errors.New("повторное использование refresh-токена")
)

type SessionFields struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	RefreshTokenHash    string
	PreviousRefreshHash string
	Device              string
	CreatedAt           time.Time
	ExpiresAt           time.Time
	RevokedAt           *time.Time
}

type Session struct {
	id                  uuid.UUID
	userID              uuid.UUID
	refreshTokenHash    string
	previousRefreshHash string
	device              string
	createdAt           time.Time
	expiresAt           time.Time
	revokedAt           *time.Time
}

func (session *Session) ID() uuid.UUID {
	return session.id
}

func (session *Session) UserID() uuid.UUID {
	return session.userID
}

func (session *Session) RefreshTokenHash() string {
	return session.refreshTokenHash
}

func (session *Session) PreviousRefreshHash() string {
	return session.previousRefreshHash
}

func (session *Session) Device() string {
	return session.device
}

func (session *Session) CreatedAt() time.Time {
	return session.createdAt
}

func (session *Session) ExpiresAt() time.Time {
	return session.expiresAt
}

func (session *Session) RevokedAt() (time.Time, bool) {
	if session.revokedAt == nil {
		return time.Time{}, false
	}

	return *session.revokedAt, true
}

func NewSession(id, userID uuid.UUID, refreshTokenHash, device string, now time.Time, ttl time.Duration) (*Session, error) {
	if ttl <= 0 {
		return nil, ErrInvalidTTL
	}

	return LoadSession(SessionFields{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		Device:           device,
		CreatedAt:        now,
		ExpiresAt:        now.Add(ttl),
	})
}

func LoadSession(fields SessionFields) (*Session, error) {
	if fields.ID == uuid.Nil() {
		return nil, ErrInvalidSessionID
	}

	if fields.UserID == uuid.Nil() {
		return nil, ErrInvalidUserID
	}

	if fields.RefreshTokenHash == "" || fields.RefreshTokenHash == fields.PreviousRefreshHash {
		return nil, ErrInvalidTokenHash
	}

	if !fields.ExpiresAt.After(fields.CreatedAt) {
		return nil, ErrInvalidSessionTime
	}

	if fields.RevokedAt != nil && fields.RevokedAt.Before(fields.CreatedAt) {
		return nil, ErrInvalidSessionTime
	}

	return &Session{
		id:                  fields.ID,
		userID:              fields.UserID,
		refreshTokenHash:    fields.RefreshTokenHash,
		previousRefreshHash: fields.PreviousRefreshHash,
		device:              normalizeDevice(fields.Device),
		createdAt:           fields.CreatedAt,
		expiresAt:           fields.ExpiresAt,
		revokedAt:           copyTime(fields.RevokedAt),
	}, nil
}

func normalizeDevice(device string) string {
	device = strings.TrimSpace(device)

	if utf8.RuneCountInString(device) > maxDeviceLength {
		device = string([]rune(device)[:maxDeviceLength])
	}

	return device
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := *value

	return &copied
}

func (session *Session) checkActive(now time.Time) error {
	if session.revokedAt != nil {
		return ErrSessionRevoked
	}

	if !now.Before(session.expiresAt) {
		return ErrSessionExpired
	}

	return nil
}

func (session *Session) Revoke(now time.Time) {
	if session.revokedAt != nil {
		return
	}

	if now.Before(session.createdAt) {
		now = session.createdAt
	}

	session.revokedAt = &now
}

func (session *Session) Refresh(presentedHash, newHash string, now time.Time, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrInvalidTTL
	}

	if newHash == "" || newHash == session.refreshTokenHash || newHash == session.previousRefreshHash {
		return ErrInvalidTokenHash
	}

	err := session.checkActive(now)
	if err != nil {
		return err
	}

	if session.previousRefreshHash != "" && presentedHash == session.previousRefreshHash {
		session.Revoke(now)
		return ErrRefreshTokenReused
	}

	if presentedHash != session.refreshTokenHash {
		return ErrInvalidRefreshToken
	}

	session.previousRefreshHash = session.refreshTokenHash
	session.refreshTokenHash = newHash
	session.expiresAt = now.Add(ttl)

	return nil
}

func (session *Session) AccessClaims(now time.Time, ttl time.Duration) (AccessClaims, error) {
	if ttl <= 0 {
		return AccessClaims{}, ErrInvalidTTL
	}

	err := session.checkActive(now)
	if err != nil {
		return AccessClaims{}, err
	}

	expiresAt := now.Add(ttl)
	if expiresAt.After(session.expiresAt) {
		expiresAt = session.expiresAt
	}

	return AccessClaims{
		SessionID: session.id,
		UserID:    session.userID,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}, nil
}
