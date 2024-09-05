package subscription

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrorAlreadyExists = errors.New("already exists")
	ErrorNotFound      = errors.New("not found")
)

const (
	insertQuery     = `INSERT INTO subscriptions (plan_code, datacenters, user_id, user) VALUES (?, ?, ?, ?);`
	selectQuery     = `SELECT plan_code, datacenters, user FROM subscriptions WHERE id = ? AND user_id = ?;`
	selectUserQuery = `SELECT id, plan_code, datacenters, user FROM subscriptions WHERE user_id = ?;`
	listQuery       = `SELECT id, plan_code, datacenters, user FROM subscriptions ORDER BY %s DESC LIMIT ? OFFSET ?;`
	listQueryCount  = `SELECT count(*) FROM subscriptions;`
	deleteQuery     = `DELETE FROM subscriptions WHERE id = ? AND user_id = ?;`
)

type Database struct {
	*sql.DB
}

func NewDatabase(filename string) (*Database, error) {
	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{},
	)

	db, err := sql.Open("sqlite3_extended", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func (db Database) Insert(s Subscription) (int64, error) {
	userJSON, err := json.Marshal(s.User)
	if err != nil {
		return -1, err
	}

	r, err := db.DB.Exec(insertQuery, s.PlanCode, strings.Join(s.Datacenters, ","), s.User.ID, string(userJSON))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			//if errors.Is(err, sqlite3.ErrConstraintUnique) {
			return -1, fmt.Errorf("%w %w", ErrorAlreadyExists, err)
		}

		return -1, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (db Database) Query(id int64, user_id int64) (*Subscription, error) {
	var s = &Subscription{}
	var datacenters string
	var userJSON string

	row := db.DB.QueryRow(selectQuery, id, user_id)
	err := row.Scan(&s.PlanCode, &datacenters, &userJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %w", ErrorNotFound, err)
		}

		return nil, err
	}

	err = json.Unmarshal([]byte(userJSON), &s.User)
	if err != nil {
		return nil, err
	}

	s.Datacenters = strings.Split(datacenters, ",")

	return s, nil
}

func (db Database) QueryUser(user_id int64) (map[int64]Subscription, error) {
	rows, err := db.DB.Query(selectUserQuery, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions := make(map[int64]Subscription)
	for rows.Next() {
		var s = Subscription{}
		var id int64
		var datacenters string
		var userJSON string

		err = rows.Scan(&id, &s.PlanCode, &datacenters, &userJSON)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(userJSON), &s.User)
		if err != nil {
			return nil, err
		}

		s.Datacenters = strings.Split(datacenters, ",")

		subscriptions[id] = s
	}

	return subscriptions, nil
}

func (db Database) QueryList(sortBy string, limit, offset int) (map[int64]map[int64]Subscription, int, error) {
	var dbQuery, dbQueryCount string

	sortBy = url.QueryEscape(sortBy)
	dbQuery = fmt.Sprintf(listQuery, sortBy)
	dbQueryCount = listQueryCount

	rows, err := db.DB.Query(dbQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var count int
	err = db.DB.QueryRow(dbQueryCount).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	subscriptions := make(map[int64]map[int64]Subscription)
	for rows.Next() {
		var s = Subscription{}
		var id int64
		var datacenters string
		var userJSON string

		err = rows.Scan(&id, &s.PlanCode, &datacenters, &userJSON)
		if err != nil {
			return nil, 0, err
		}

		err = json.Unmarshal([]byte(userJSON), &s.User)
		if err != nil {
			return nil, 0, err
		}

		s.Datacenters = strings.Split(datacenters, ",")

		if _, ok := subscriptions[s.User.ID]; !ok {
			subscriptions[s.User.ID] = make(map[int64]Subscription)
		}

		subscriptions[s.User.ID][id] = s
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return subscriptions, count, nil
}

func (db Database) Delete(id int64, user_id int64) error {
	_, err := db.DB.Exec(deleteQuery, id, user_id)
	if err != nil {
		return err
	}
	return nil
}
