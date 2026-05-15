package domain

import "github.com/jackc/pgx/v5/pgtype"

type Project struct {
	Id     pgtype.UUID `json:"id"`
	Name   string      `json:"name"`
	UserId pgtype.UUID `json:"userId"`
}
