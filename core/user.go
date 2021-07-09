package core

import (
	"context"

	"github.com/tomekwlod/booking/internal/db"
)

type User struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

type UserStore interface {
	Create(context.Context, db.Execer, *User) error
}
