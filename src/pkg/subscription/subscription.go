package subscription

import (
	"time"

	tele "gopkg.in/telebot.v3"
)

type Subscription struct {
	PlanCode    string
	Datacenters []string
	User        *tele.User
	LastCheck   time.Time
}

type Service struct {
	Database *Database
}

func NewService(databaseFilename string) (s *Service, err error) {
	s = &Service{}

	s.Database, err = NewDatabase(databaseFilename)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Subscribe(telegramUser *tele.User, planCode string, datacenters []string) (int64, error) {
	subscription := Subscription{
		PlanCode:    planCode,
		Datacenters: datacenters,
		User:        telegramUser,
		LastCheck:   time.Now(),
	}

	id, err := s.Database.Insert(subscription)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *Service) Unsubscribe(telegramUser *tele.User, subscriptionId int64) error {
	err := s.Database.Delete(subscriptionId, telegramUser.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ListUser(telegramUser *tele.User) (map[int64]Subscription, error) {
	subscriptions, err := s.Database.QueryUser(telegramUser.ID)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (s *Service) ListPaginate(sortBy string, limit, offset int) (map[int64]map[int64]Subscription, int, error) {
	subscriptions, count, err := s.Database.QueryList(sortBy, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return subscriptions, count, nil
}
