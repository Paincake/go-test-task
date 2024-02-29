package postgres

import (
	"fmt"
	"github.com/Paincake/go-test-task/internal/database"
	"github.com/Paincake/go-test-task/internal/lib/api/sort"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"strings"
)

type Database struct {
	db *sqlx.DB
}

func (d *Database) GetPerson(id int64) {
	//TODO implement me
	panic("implement me")
}

func New(dbname, username, password, host, port string) (*Database, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?&sslmode=disable",
		username,
		password,
		host,
		port,
		dbname)
	db, err := sqlx.Connect("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &Database{db: db}, nil
}

func (d *Database) SavePerson(age int, name, surname, patronymic, gender, nationality string) (int64, error) {
	const op = "internal.database.postgres.SavePerson"
	var id int64
	err := d.db.Get(&id,
		"INSERT INTO persons (age, name, surname, patronymic, gender, nationality) VALUES ( $1, $2, $3, $4, $5, $6) RETURNING person_id;",
		age, name, surname, patronymic, gender, nationality)
	if err != nil {
		return 0, fmt.Errorf("%s: error creating INSERT into table 'persons' statement: %w", op, err)
	}

	return id, nil
}

func (d *Database) GetPersons(limit int, idCursor int64, sortOptions sort.Options, filterOptions string) ([]database.Person, int64, error) {
	const op = "internal.database.postgres.GetPerson"
	var persons []database.Person
	query := "SELECT person_id, name, surname, patronymic, age, gender, nationality" +
		" FROM persons"
	query += fmt.Sprintf("\n%s", parseFilterOptions(filterOptions, idCursor))
	query += fmt.Sprintf("\n%s %s", sortOptions.Field, sortOptions.Order)
	query += fmt.Sprintf("\n LIMIT %d", limit)

	err := d.db.Select(&persons, query)
	if err != nil {
		return nil, -1, fmt.Errorf("%s: error fetching persons from database: %w", op, err)
	}
	var cursor int64
	if len(persons) == 0 {
		cursor = idCursor
	} else {
		cursor = persons[len(persons)-1].Id
	}
	return persons, cursor, nil

}

func parseFilterOptions(filterOptions string, cursor int64) string {
	// получаем вот такую строку: age>=40&age<60&name=Egor
	filterConditions := strings.Split(filterOptions, "%26")
	whereQuery := "WHERE"
	for _, e := range filterConditions {
		if len(e) == 0 {
			continue
		}
		var filterIdx int
		var filterLen int
		var cond string
		if strings.Index(e, ">=") != -1 {
			filterIdx = strings.Index(e, ">=")
			filterLen = 2
			cond = ">="
		} else if strings.Index(e, "<=") != -1 {
			filterIdx = strings.Index(e, "<=")
			filterLen = 2
			cond = "<="
		} else if strings.Index(e, "=") != -1 {
			filterIdx = strings.Index(e, "=")
			filterLen = 1
			cond = "="
		} else if strings.Index(e, ">") != -1 {
			filterIdx = strings.Index(e, ">")
			filterLen = 1
			cond = ">"
		} else if strings.Index(e, "<") != -1 {
			filterIdx = strings.Index(e, "<")
			filterLen = 1
			cond = "<"
		}
		field := e[0 : filterIdx+1]
		value := e[filterIdx+filterLen:]
		if len(whereQuery) > 5 {
			whereQuery += " AND "
		}
		whereQuery += fmt.Sprintf(" %s %s %s", field, cond, value)
	}
	if len(whereQuery) > 5 {
		whereQuery += " AND"
	}
	whereQuery += fmt.Sprintf(" person_id > %d", cursor)
	return whereQuery
}
