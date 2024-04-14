package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/avito-tech/internal/user"
	"github.com/KretovDmitry/avito-tech/pkg/log"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID int) (*user.User, error)
}

type repository struct {
	db     *sql.DB
	logger log.Logger
}

func NewRepository(db *sql.DB, logger log.Logger) (*repository, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}

	return &repository{
		db:     db,
		logger: logger,
	}, nil
}

var _ Repository = (*repository)(nil)

func (r *repository) GetUserByID(ctx context.Context, userID int) (*user.User, error) {
	const query = `
			SELECT
				id, name, role, created_at, updated_at
			FROM
				users
			WHERE
				id = $1
	`

	row := r.db.QueryRowContext(ctx, query, userID)
	var u user.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return &u, nil
}
