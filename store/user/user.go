package user

import (
	"context"

	"github.com/tomekwlod/booking/core"
	"github.com/tomekwlod/booking/internal/db"
)

// New returns a new UserStore.
func New(db *db.DB) core.UserStore {
	return &userStore{db}
}

type userStore struct {
	db *db.DB
}

func (r *userStore) Create(ctx context.Context, db db.Execer, u *core.User) error { // User instead of the interface{}

	// postgres has $1 binding whereas mysql has ?
	return db.QueryRowContext(ctx, `INSERT INTO users (password, email) 
    VALUES ($1,$2) RETURNING id`, u.Password, u.Email).Scan(&u.ID)
}
