package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(port, user, password, name, host string) (*Storage, error) {
	const op = "storage.postgres.New"
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil

}
func (s *Storage) UserBalance(user string) (int, error) {
	const op = "storage.postgres.UserBalance"
	qBuilder := squirrel.Select("coins").From("users").Where("name=?", user)
	query, args, err := qBuilder.ToSql()
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	var balance int
	err = s.db.QueryRow(query, args...).Scan(&balance)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	return balance, nil
}
