package subscription

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	tele "gopkg.in/telebot.v3"
)

var (
	ErrorAlreadyExists = errors.New("already exists")
	ErrorNotFound      = errors.New("not found")
)

const (
	listQueryCount                           = `SELECT count(*) FROM subscriptions;`
	updateNotificationsLastCheckSubscription = `UPDATE subscriptions SET last_check = ?, notifications = ? WHERE id = ?;`

	insertSubscription = `INSERT OR IGNORE INTO users (user_id, user) VALUES (?, ?);` +
		`INSERT OR IGNORE INTO subscriptions (plan_code, datacenters, region, last_check) VALUES (?, ?, ?, ?);` +
		//`INSERT INTO user_subscriptions (user_id, subscription_id) VALUES(?, 12);`
		`INSERT INTO user_subscriptions (user_id, subscription_id) VALUES(?, (SELECT id FROM subscriptions WHERE plan_code = ? AND datacenters = ? AND region = ?));`

	selectUserSubscriptions = `SELECT us.id, s.plan_code, s.datacenters, s.region, s.last_check, s.notifications FROM user_subscriptions AS us JOIN subscriptions AS s ON us.subscription_id = s.id WHERE us.user_id = ?;`

	listSubscriptions               = `SELECT s.id, s.plan_code, s.datacenters, s.region, s.last_check, s.notifications, u.user FROM user_subscriptions AS us JOIN subscriptions AS s ON us.subscription_id = s.id JOIN users AS u ON us.user_id = u.user_id ORDER BY us.id ASC LIMIT ? OFFSET ?;`
	countSubscriptions              = `SELECT count(*) FROM user_subscriptions;`
	deleteUserSubscription          = `DELETE FROM user_subscriptions WHERE subscription_id = ? AND user_id = ?;`
	deleteAllUserSubscriptions      = `DELETE FROM user_subscriptions WHERE user_id = ?;`
	deleteMultipleSubscriptions     = `DELETE FROM user_subscriptions WHERE subscription_id = ? AND user_id IN(?);`
	updateSubscriptionNotifications = `UPDATE subscriptions SET notifications = notifications ? WHERE id = ?;`
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

// InsertSubscription replaces Insert
func (db Database) InsertSubscription(s Subscription, user *tele.User) (int64, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return -1, err
	}

	r, err := db.DB.Exec(insertSubscription,
		user.ID,
		string(userJSON),
		s.PlanCode,
		strings.Join(s.Datacenters, ","),
		s.Region,
		s.LastCheck.Format(time.RFC3339),
		user.ID,
		s.PlanCode,
		strings.Join(s.Datacenters, ","),
		s.Region,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			//if errors.Is(err, sqlite3.ErrConstraintUnique) {
			return -1, fmt.Errorf("%w %w", ErrorAlreadyExists, err)
		}

		return -1, err
	}

	// Is this the subscriptionId?
	id, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

// QueryUserSubscriptions replaces QueryUser
func (db Database) QueryUserSubscriptions(user_id int64) ([]Subscription, error) {
	rows, err := db.DB.Query(selectUserSubscriptions, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %w", ErrorNotFound, err)
		}

		return nil, err
	}
	defer rows.Close()

	subscriptions := make([]Subscription, 0)
	for rows.Next() {
		var s = Subscription{}
		var datacenters string
		var lastCheckString string

		err = rows.Scan(&s.ID, &s.PlanCode, &datacenters, &s.Region, &lastCheckString, &s.Notifications)
		if err != nil {
			return nil, err
		}

		s.LastCheck, err = time.Parse(time.RFC3339, lastCheckString)
		if err != nil {
			return nil, err
		}

		s.Datacenters = strings.Split(datacenters, ",")

		subscriptions = append(subscriptions, s)
	}

	return subscriptions, nil
}

func (db Database) QueryList(sortBy string, limit, offset int) (map[int64]Subscription, int, error) {
	rows, err := db.DB.Query(listSubscriptions, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("%w %w", ErrorNotFound, err)
		}
		return nil, 0, err
	}
	defer rows.Close()

	var count int
	err = db.DB.QueryRow(listQueryCount).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	subscriptions := make(map[int64]Subscription)
	for rows.Next() {
		var s = Subscription{}
		var datacenters string
		var lastCheckString string
		var userJSON string
		var user *tele.User

		err = rows.Scan(&s.ID, &s.PlanCode, &datacenters, &s.Region, &lastCheckString, &s.Notifications, &userJSON)
		if err != nil {
			return nil, 0, err
		}

		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			return nil, 0, err
		}

		s.LastCheck, err = time.Parse(time.RFC3339, lastCheckString)
		if err != nil {
			return nil, 0, err
		}

		s.Datacenters = strings.Split(datacenters, ",")

		s.Users = append(s.Users, user)
		subscriptions[s.ID] = s
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return subscriptions, count, nil
}

func (db Database) Delete(id int64, user_id int64) error {
	r, err := db.DB.Exec(deleteUserSubscription, id, user_id)
	if err != nil {
		return err
	}

	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w %w", ErrorNotFound, errors.New("no rows affected"))
	}

	return nil
}

func (db Database) DeleteAll(user_id int64) error {
	r, err := db.DB.Exec(deleteAllUserSubscriptions, user_id)
	if err != nil {
		return err
	}

	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w %w", ErrorNotFound, errors.New("no rows affected"))
	}

	return nil
}

func (db Database) DeleteMultiple(subscription_id int64, userIDs []int64) error {
	var ids []string
	for _, id := range userIDs {
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	r, err := db.DB.Exec(deleteMultipleSubscriptions, subscription_id, strings.Join(ids, ","))
	if err != nil {
		return err
	}

	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w %w", ErrorNotFound, errors.New("no rows affected"))
	}

	return nil
}

func (db Database) UpdateNotificationsLastCheck(subscription Subscription) error {
	_, err := db.DB.Exec(updateNotificationsLastCheckSubscription, time.Now().Format(time.RFC3339), subscription.Notifications, subscription.ID)
	if err != nil {
		return err
	}

	return nil
}
