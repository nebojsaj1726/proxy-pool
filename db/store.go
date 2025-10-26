package db

import "database/sql"

type UserStore interface {
	CreateUser(id, username, passwordHash string) error
	GetUserByUsername(username string) (id string, passwordHash string, err error)
}

type Store struct {
	DB *sql.DB
}
