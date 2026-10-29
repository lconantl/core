package domain_test

import (
	"session/internal/session/domain"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
)

const (
	refreshTTL = 30 * 24 * time.Hour
	accessTTL  = 15 * time.Minute

	testDevice = "Chrome on Windows"
	firstHash  = "hash-1"
	secondHash = "hash-2"
	thirdHash  = "hash-3"
)

var testNow = time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)

func newSession(test *testing.T) *domain.Session {
	test.Helper()

	session, err := domain.NewSession(uuid.New(), uuid.New(), firstHash, testDevice, testNow, refreshTTL)
	require.NoError(test, err)

	return session
}

func validFields() domain.SessionFields {
	return domain.SessionFields{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		RefreshTokenHash: firstHash,
		Device:           testDevice,
		CreatedAt:        testNow,
		ExpiresAt:        testNow.Add(refreshTTL),
	}
}

func TestNewSession(test *testing.T) {
	id := uuid.New()
	userID := uuid.New()

	session, err := domain.NewSession(id, userID, firstHash, "  "+testDevice+"  ", testNow, refreshTTL)

	require.NoError(test, err)
	require.Equal(test, id, session.ID())
	require.Equal(test, userID, session.UserID())
	require.Equal(test, firstHash, session.RefreshTokenHash())
	require.Empty(test, session.PreviousRefreshHash(), "у новой сессии нет предыдущего токена")
	require.Equal(test, testDevice, session.Device(), "device обрезается по trim")
	require.Equal(test, testNow, session.CreatedAt())
	require.Equal(test, testNow.Add(refreshTTL), session.ExpiresAt())

	_, revoked := session.RevokedAt()
	require.False(test, revoked)
}

func TestDeviceTooLong(test *testing.T) {
	session, err := domain.NewSession(uuid.New(), uuid.New(), firstHash, strings.Repeat("я", 300), testNow, refreshTTL)

	require.NoError(test, err, "длинный device — не ошибка")
	require.Equal(test, strings.Repeat("я", 255), session.Device())
}

type newSessionCase struct {
	name   string
	id     uuid.UUID
	userID uuid.UUID
	hash   string
	ttl    time.Duration
	err    error
}

func checkNewSession(test *testing.T, testCase newSessionCase) {
	test.Helper()

	session, err := domain.NewSession(testCase.id, testCase.userID, testCase.hash, testDevice, testNow, testCase.ttl)

	require.ErrorIs(test, err, testCase.err)
	require.Nil(test, session)
}

func TestNewSessionErrors(test *testing.T) {
	testCases := []newSessionCase{
		{"пустой ID", uuid.Nil(), uuid.New(), firstHash, refreshTTL, domain.ErrInvalidSessionID},
		{"пустой userID", uuid.New(), uuid.Nil(), firstHash, refreshTTL, domain.ErrInvalidUserID},
		{"пустой хеш", uuid.New(), uuid.New(), "", refreshTTL, domain.ErrInvalidTokenHash},
		{"нулевой TTL", uuid.New(), uuid.New(), firstHash, 0, domain.ErrInvalidTTL},
		{"отрицательный TTL", uuid.New(), uuid.New(), firstHash, -time.Hour, domain.ErrInvalidTTL},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkNewSession(test, testCase)
		})
	}
}

type loadSessionCase struct {
	name   string
	modify func(fields *domain.SessionFields)
	err    error
}

func checkLoadSession(test *testing.T, testCase loadSessionCase) {
	test.Helper()

	fields := validFields()
	testCase.modify(&fields)

	session, err := domain.LoadSession(fields)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
		require.Nil(test, session)
		return
	}

	require.NoError(test, err)
	require.Equal(test, fields.ID, session.ID())
	require.Equal(test, fields.RefreshTokenHash, session.RefreshTokenHash())
	require.Equal(test, fields.PreviousRefreshHash, session.PreviousRefreshHash())
	require.Equal(test, fields.ExpiresAt, session.ExpiresAt())

	revokedAt, revoked := session.RevokedAt()
	require.Equal(test, fields.RevokedAt != nil, revoked)
	if fields.RevokedAt != nil {
		require.Equal(test, *fields.RevokedAt, revokedAt)
	}
}

func TestLoadSession(test *testing.T) {
	later := testNow.Add(time.Hour)
	earlier := testNow.Add(-time.Hour)

	testCases := []loadSessionCase{
		{"активная", func(*domain.SessionFields) {}, nil},
		{"после ротации", func(fields *domain.SessionFields) { fields.PreviousRefreshHash = secondHash }, nil},
		{"отозванная", func(fields *domain.SessionFields) { fields.RevokedAt = &later }, nil},
		{"пустой ID", func(fields *domain.SessionFields) { fields.ID = uuid.Nil() }, domain.ErrInvalidSessionID},
		{"пустой userID", func(fields *domain.SessionFields) { fields.UserID = uuid.Nil() }, domain.ErrInvalidUserID},
		{"пустой хеш", func(fields *domain.SessionFields) { fields.RefreshTokenHash = "" }, domain.ErrInvalidTokenHash},
		{"текущий хеш равен предыдущему", func(fields *domain.SessionFields) { fields.PreviousRefreshHash = firstHash }, domain.ErrInvalidTokenHash},
		{"expiresAt равен createdAt", func(fields *domain.SessionFields) { fields.ExpiresAt = testNow }, domain.ErrInvalidSessionTime},
		{"expiresAt раньше createdAt", func(fields *domain.SessionFields) { fields.ExpiresAt = earlier }, domain.ErrInvalidSessionTime},
		{"revokedAt раньше createdAt", func(fields *domain.SessionFields) { fields.RevokedAt = &earlier }, domain.ErrInvalidSessionTime},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkLoadSession(test, testCase)
		})
	}
}

func TestLoadSessionCopiesRevokedAt(test *testing.T) {
	fields := validFields()
	revokedAt := testNow.Add(time.Hour)
	fields.RevokedAt = &revokedAt

	session, err := domain.LoadSession(fields)
	require.NoError(test, err)

	revokedAt = revokedAt.Add(time.Hour)

	stored, _ := session.RevokedAt()
	require.Equal(test, testNow.Add(time.Hour), stored, "изменение исходных данных не влияет на сессию")
}

func TestRevoke(test *testing.T) {
	session := newSession(test)

	session.Revoke(testNow)
	session.Revoke(testNow.Add(time.Hour))

	revokedAt, revoked := session.RevokedAt()
	require.True(test, revoked)
	require.Equal(test, testNow, revokedAt, "повторный Revoke не перезаписывает дату")
}

func TestRevokeExpired(test *testing.T) {
	session := newSession(test)
	afterExpiry := testNow.Add(refreshTTL + time.Hour)

	session.Revoke(afterExpiry)

	revokedAt, revoked := session.RevokedAt()
	require.True(test, revoked, "истёкшую сессию можно отозвать")
	require.Equal(test, afterExpiry, revokedAt)
}

func TestRefresh(test *testing.T) {
	session := newSession(test)
	later := testNow.Add(time.Hour)

	err := session.Refresh(firstHash, secondHash, later, refreshTTL)

	require.NoError(test, err)
	require.Equal(test, secondHash, session.RefreshTokenHash())
	require.Equal(test, firstHash, session.PreviousRefreshHash())
	require.Equal(test, later.Add(refreshTTL), session.ExpiresAt(), "сессия продлевается от момента refresh")

	_, err = session.AccessClaims(later, accessTTL)
	require.NoError(test, err, "после refresh сессия активна")
}

func TestRefreshTwice(test *testing.T) {
	session := newSession(test)

	require.NoError(test, session.Refresh(firstHash, secondHash, testNow, refreshTTL))
	require.NoError(test, session.Refresh(secondHash, thirdHash, testNow, refreshTTL))

	require.Equal(test, thirdHash, session.RefreshTokenHash())
	require.Equal(test, secondHash, session.PreviousRefreshHash(), "хранится только последний предыдущий")
}

func TestRefreshTokenReuse(test *testing.T) {
	session := newSession(test)
	later := testNow.Add(time.Hour)
	require.NoError(test, session.Refresh(firstHash, secondHash, testNow, refreshTTL))

	err := session.Refresh(firstHash, thirdHash, later, refreshTTL)

	require.ErrorIs(test, err, domain.ErrRefreshTokenReused)

	revokedAt, revoked := session.RevokedAt()
	require.True(test, revoked, "предъявлен старый токен — сессия отозвана")
	require.Equal(test, later, revokedAt)
	require.Equal(test, secondHash, session.RefreshTokenHash(), "хеш не ротируется")
	require.ErrorIs(test, session.Refresh(secondHash, thirdHash, later, refreshTTL), domain.ErrSessionRevoked,
		"после кражи не работает и актуальный токен")
}

type refreshCase struct {
	name      string
	revoke    bool
	presented string
	newHash   string
	now       time.Time
	ttl       time.Duration
	err       error
}

func checkRefresh(test *testing.T, testCase refreshCase) {
	test.Helper()

	session := newSession(test)
	if testCase.revoke {
		session.Revoke(testNow)
	}

	err := session.Refresh(testCase.presented, testCase.newHash, testCase.now, testCase.ttl)

	require.ErrorIs(test, err, testCase.err)
	require.Equal(test, firstHash, session.RefreshTokenHash(), "при ошибке хеш не меняется")
	require.Empty(test, session.PreviousRefreshHash(), "при ошибке предыдущий хеш не меняется")
	require.Equal(test, testNow.Add(refreshTTL), session.ExpiresAt(), "при ошибке срок не меняется")
}

func TestRefreshErrors(test *testing.T) {
	expired := testNow.Add(refreshTTL)

	testCases := []refreshCase{
		{"чужой токен", false, "unknown", secondHash, testNow, refreshTTL, domain.ErrInvalidRefreshToken},
		{"пустой предъявленный токен", false, "", secondHash, testNow, refreshTTL, domain.ErrInvalidRefreshToken},
		{"сессия истекла", false, firstHash, secondHash, expired, refreshTTL, domain.ErrSessionExpired},
		{"сессия отозвана", true, firstHash, secondHash, testNow, refreshTTL, domain.ErrSessionRevoked},
		{"пустой новый хеш", false, firstHash, "", testNow, refreshTTL, domain.ErrInvalidTokenHash},
		{"новый хеш равен текущему", false, firstHash, firstHash, testNow, refreshTTL, domain.ErrInvalidTokenHash},
		{"нулевой TTL", false, firstHash, secondHash, testNow, 0, domain.ErrInvalidTTL},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkRefresh(test, testCase)
		})
	}
}

func TestAccessClaims(test *testing.T) {
	session := newSession(test)

	claims, err := session.AccessClaims(testNow, accessTTL)

	require.NoError(test, err)
	require.Equal(test, session.ID(), claims.SessionID)
	require.Equal(test, session.UserID(), claims.UserID)
	require.Equal(test, testNow, claims.IssuedAt)
	require.Equal(test, testNow.Add(accessTTL), claims.ExpiresAt)
}

func TestAccessClaimsNotOutliveSession(test *testing.T) {
	session := newSession(test)
	almostExpired := session.ExpiresAt().Add(-time.Minute)

	claims, err := session.AccessClaims(almostExpired, accessTTL)

	require.NoError(test, err)
	require.Equal(test, session.ExpiresAt(), claims.ExpiresAt, "access не живёт дольше сессии")
}

type accessClaimsCase struct {
	name   string
	revoke bool
	now    time.Time
	ttl    time.Duration
	err    error
}

func checkAccessClaims(test *testing.T, testCase accessClaimsCase) {
	test.Helper()

	session := newSession(test)
	if testCase.revoke {
		session.Revoke(testNow)
	}

	_, err := session.AccessClaims(testCase.now, testCase.ttl)

	if testCase.err != nil {
		require.ErrorIs(test, err, testCase.err)
		return
	}

	require.NoError(test, err)
}

func TestAccessClaimsActivity(test *testing.T) {
	expiresAt := testNow.Add(refreshTTL)

	testCases := []accessClaimsCase{
		{"только что создана", false, testNow, accessTTL, nil},
		{"за наносекунду до истечения", false, expiresAt.Add(-time.Nanosecond), accessTTL, nil},
		{"ровно в момент истечения", false, expiresAt, accessTTL, domain.ErrSessionExpired},
		{"после истечения", false, expiresAt.Add(time.Hour), accessTTL, domain.ErrSessionExpired},
		{"отозвана", true, testNow, accessTTL, domain.ErrSessionRevoked},
		{"отозвана и истекла — важнее отзыв", true, expiresAt.Add(time.Hour), accessTTL, domain.ErrSessionRevoked},
		{"нулевой TTL", false, testNow, 0, domain.ErrInvalidTTL},
		{"отрицательный TTL", false, testNow, -time.Minute, domain.ErrInvalidTTL},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			checkAccessClaims(test, testCase)
		})
	}
}
