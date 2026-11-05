package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1794646184(client mysql.ClientContext) migrator.Migration {
	return &version1794646184{
		client: client,
	}
}

type version1794646184 struct {
	client mysql.ClientContext
}

func (v *version1794646184) Version() int64 {
	return 1794646184
}

func (v *version1794646184) Description() string {
	return "Create credentials table"
}

func (v *version1794646184) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-credentials-table", `
CREATE TABLE IF NOT EXISTS credentials (
    user_id CHAR(36) CHARACTER SET ascii NOT NULL,
    username VARCHAR(32) CHARACTER SET ascii NOT NULL,
    password_hash VARCHAR(255) CHARACTER SET ascii NOT NULL,

    PRIMARY KEY (user_id),
    UNIQUE KEY uq_credentials_username (username)
)
    ENGINE=InnoDB
    CHARACTER SET=utf8mb4
    COLLATE=utf8mb4_unicode_ci
`)
	return err
}
