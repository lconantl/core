package repository

import (
	"auth/internal/auth/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	mysqldriver "github.com/go-sql-driver/mysql"
)

const usernameUniqueKey = "uq_credentials_username"

func NewCredentialsRepository(ctx context.Context, client mysql.ClientContext) domain.CredentialsRepository {
	return &credentialsRepository{
		ctx:    ctx,
		client: client,
	}
}

type credentialsRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type credentialsRow struct {
	UserID       string `db:"user_id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
}

func (r *credentialsRepository) Add(credentials *domain.Credentials) error {
	_, err := r.client.ExecContext(r.ctx, "credentials.add", `INSERT INTO credentials (user_id, username, password_hash) VALUES (?, ?, ?)`, credentials.UserID().String(), credentials.Username().String(), credentials.PasswordHash())

	if isDuplicateKey(err, usernameUniqueKey) {
		return domain.ErrUsernameTaken
	}

	return err
}

func (r *credentialsRepository) Update(credentials *domain.Credentials) error {
	_, err := r.client.ExecContext(r.ctx, "credentials.update",
		`UPDATE credentials SET password_hash = ? WHERE user_id = ?`,
		credentials.PasswordHash(), credentials.UserID().String(),
	)

	return err
}

func (r *credentialsRepository) FindByUsername(username domain.Username) (*domain.Credentials, error) {
	return r.find("credentials.find-by-username", `SELECT user_id, username, password_hash FROM credentials`+` WHERE username = ?`, username.String())
}

func (r *credentialsRepository) FindByUserIDForUpdate(userID uuid.UUID) (*domain.Credentials, error) {
	return r.find("credentials.find-by-user-id-for-update", `SELECT user_id, username, password_hash FROM credentials`+` WHERE user_id = ? FOR UPDATE`, userID.String())
}

func (r *credentialsRepository) find(queryID, query, arg string) (*domain.Credentials, error) {
	var row credentialsRow

	err := r.client.GetContext(r.ctx, queryID, &row, query, arg)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCredentialsNotFound
	}

	if err != nil {
		return nil, err
	}

	return toCredentials(row)
}

func toCredentials(row credentialsRow) (*domain.Credentials, error) {
	userID, err := uuid.Parse(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("credentials %q: user id: %w", row.UserID, err)
	}

	username, err := domain.NewUsername(row.Username)
	if err != nil {
		return nil, fmt.Errorf("credentials %q: username: %w", row.UserID, err)
	}

	return domain.LoadCredentials(userID, username, row.PasswordHash)
}

func isDuplicateKey(err error, key string) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, key)
}
