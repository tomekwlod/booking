package core

import (
	"context"
	"time"

	"github.com/tomekwlod/booking/internal/db"
)

type UserStore interface {
	Create(context.Context, db.Execer, *User) error
}

type User struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	FN        string     `json:"fn"`
	LN        string     `json:"ln"`
	Enabled   bool       `json:"enabled"`
	LastLogin *time.Time `json:"last_login"`
	// UpdatedAt string `json:"updated_at"`
	// UpdatedAt string `json:"updated_at"`
}
