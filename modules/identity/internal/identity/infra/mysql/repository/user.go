package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	mysqldriver "github.com/go-sql-driver/mysql"

	"identity/internal/identity/domain"
)

const emailUniqueKey = "uq_users_email"

func NewUserRepository(ctx context.Context, client mysql.ClientContext) domain.UserRepository {
	return &userRepository{
		ctx:    ctx,
		client: client,
	}
}

type userRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type userRow struct {
	ID           string         `db:"id"`
	Status       string         `db:"status"`
	FirstName    string         `db:"first_name"`
	LastName     string         `db:"last_name"`
	Email        string         `db:"email"`
	Phone        string         `db:"phone"`
	Description  string         `db:"description"`
	LanguageCode string         `db:"language_code"`
	AvatarID     sql.NullString `db:"avatar_id"`
}

func (r *userRepository) Add(user *domain.User) error {
	row := toUserRow(user)
	_, err := r.client.ExecContext(r.ctx, "user.add", `INSERT INTO users (id, status, first_name, last_name, email, phone, description, language_code, avatar_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.ID, row.Status, row.FirstName, row.LastName, row.Email, row.Phone, row.Description, row.LanguageCode, row.AvatarID)
	return mapWriteError(err)
}

func (r *userRepository) Update(user *domain.User) error {
	row := toUserRow(user)
	_, err := r.client.ExecContext(r.ctx, "user.update", `UPDATE users SET status = ?, first_name = ?, last_name = ?, email = ?, phone = ?, description = ?, language_code = ?, avatar_id = ? WHERE id = ?`, row.Status, row.FirstName, row.LastName, row.Email, row.Phone, row.Description, row.LanguageCode, row.AvatarID, row.ID)
	return mapWriteError(err)
}

func (r *userRepository) Delete(id uuid.UUID) error {
	_, err := r.client.ExecContext(r.ctx, "user.delete", `DELETE FROM users WHERE id = ?`, id.String())
	return err
}

func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	return r.find("user.find-by-id", `SELECT id, status, first_name, last_name, email, phone, description, language_code, avatar_id FROM users WHERE id = ?`, id)
}

func (r *userRepository) FindByIDForUpdate(id uuid.UUID) (*domain.User, error) {
	return r.find("user.find-by-id-for-update", `SELECT id, status, first_name, last_name, email, phone, description, language_code, avatar_id FROM users WHERE id = ?`+" FOR UPDATE", id)
}

func (r *userRepository) find(queryID, query string, id uuid.UUID) (*domain.User, error) {
	var row userRow

	err := r.client.GetContext(r.ctx, queryID, &row, query, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return toUser(row)
}

func mapWriteError(err error) error {
	if isDuplicateKey(err, emailUniqueKey) {
		return domain.ErrEmailTaken
	}

	return err
}

func toUserRow(user *domain.User) userRow {
	profile := user.Profile()

	row := userRow{
		ID:           user.ID().String(),
		Status:       string(user.Status()),
		FirstName:    profile.FirstName(),
		LastName:     profile.LastName(),
		Email:        profile.Email(),
		Phone:        profile.Phone(),
		Description:  profile.Description(),
		LanguageCode: profile.LanguageCode(),
	}

	if profile.HasAvatar() {
		row.AvatarID = sql.NullString{String: profile.AvatarID().String(), Valid: true}
	}

	return row
}

func toUser(row userRow) (*domain.User, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("user %q: id: %w", row.ID, err)
	}

	avatarID := uuid.Nil()
	if row.AvatarID.Valid {
		avatarID, err = uuid.Parse(row.AvatarID.String)
		if err != nil {
			return nil, fmt.Errorf("user %q: avatar id: %w", row.ID, err)
		}
	}

	profile, err := domain.NewProfile(domain.ProfileFields{
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		Email:        row.Email,
		Phone:        row.Phone,
		Description:  row.Description,
		LanguageCode: row.LanguageCode,
		AvatarID:     avatarID,
	})
	if err != nil {
		return nil, fmt.Errorf("user %q: profile: %w", row.ID, err)
	}

	return domain.LoadUser(id, domain.UserStatus(row.Status), profile)
}

func isDuplicateKey(err error, key string) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, key)
}
