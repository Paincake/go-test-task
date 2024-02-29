package database

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrUrlExists   = errors.New("url exists")
)

type Person struct {
	Id          int64 `db:"person_id"`
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality string
}
