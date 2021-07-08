package user

import (
	"context"
	"database/sql"

	"github.com/tomekwlod/booking/model"
)

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
type Execer interface {
	Querier
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
type Binder interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
}

type UserRepo struct{}

func (r *UserRepo) Create(ctx context.Context, db Execer, u *model.User) error { // User instead of the interface{}
	// postgres has $1 binding whereas mysql has ?
	return db.QueryRowContext(ctx, `INSERT INTO users (username, email, description) VALUES ($1,$2,$3) RETURNING id`, u.Username, u.Email, u.Description).Scan(&u.ID)
}
