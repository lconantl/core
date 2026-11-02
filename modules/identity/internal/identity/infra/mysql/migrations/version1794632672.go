package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1794632672(client mysql.ClientContext) migrator.Migration {
	return &version1794632672{
		client: client,
	}
}

type version1794632672 struct {
	client mysql.ClientContext
}

func (v *version1794632672) Version() int64 {
	return 1794632672
}

func (v *version1794632672) Description() string {
	return "Create users table"
}

func (v *version1794632672) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-users-table", `
CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) CHARACTER SET ascii NOT NULL,
	status ENUM('pending', 'active') NOT NULL,
    first_name VARCHAR(64) NOT NULL,
    last_name VARCHAR(64) NOT NULL,
	email VARCHAR(254) NOT NULL,
    phone VARCHAR(16) CHARACTER SET ascii NOT NULL DEFAULT '',
	description VARCHAR(70) NOT NULL DEFAULT '',
    language_code VARCHAR(2) CHARACTER SET ascii NOT NULL DEFAULT '',
    avatar_id CHAR(36) CHARACTER SET ascii NULL,

    PRIMARY KEY (id),
	UNIQUE KEY uq_users_email (email)
) 
    ENGINE=InnoDB
    CHARACTER SET=utf8mb4
    COLLATE=utf8mb4_unicode_ci
`)
	return err
}
