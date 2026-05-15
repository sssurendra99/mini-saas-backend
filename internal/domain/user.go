package domain

import "github.com/jackc/pgx/v5/pgtype"

type User struct {
	ID       pgtype.UUID `json:"id"`
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Password string      `json:"password"`
}
