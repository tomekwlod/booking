package user

import (
	"context"

	"github.com/tomekwlod/booking/core"
	"github.com/tomekwlod/booking/internal/db"
)

type UserRepo struct{}

func (r *UserRepo) Create(ctx context.Context, db db.Execer, u *core.User) error { // User instead of the interface{}
	// postgres has $1 binding whereas mysql has ?
	return db.QueryRowContext(ctx, `INSERT INTO users (username, email, description) VALUES ($1,$2,$3) RETURNING id`, u.Username, u.Email, u.Description).Scan(&u.ID)
}
