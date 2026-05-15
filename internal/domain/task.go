package domain

import "github.com/jackc/pgx/v5/pgtype"

type Task struct {
	ID        pgtype.UUID `json:"id"`
	Title     string      `json:"title"`
	Completed bool        `json:"completed"`
	UserId    pgtype.UUID `json:"userId"`
}
