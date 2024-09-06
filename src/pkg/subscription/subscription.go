package subscription

import (
	"time"

	tele "gopkg.in/telebot.v3"
)

type Subscription struct {
	ID            int64
	PlanCode      string
	Datacenters   []string
	Region        string
	LastCheck     time.Time
	Users         []*tele.User
	Notifications int64
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

func (s *Service) Subscribe(telegramUser *tele.User, region, planCode string, datacenters []string) (int64, error) {
	subscription := Subscription{
		PlanCode:    planCode,
		Datacenters: datacenters,
		Region:      region,
		LastCheck:   time.Now(),
	}

	id, err := s.Database.InsertSubscription(subscription, telegramUser)
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

func (s *Service) UnsubscribeAll(telegramUser *tele.User) error {
	err := s.Database.DeleteAll(telegramUser.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UnsubscribeMultiple(subscriptionId int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	err := s.Database.DeleteMultiple(subscriptionId, userIDs)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ListUser(telegramUser *tele.User) ([]Subscription, error) {
	subscriptions, err := s.Database.QueryUserSubscriptions(telegramUser.ID)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (s *Service) ListPaginate(sortBy string, limit, offset int) (map[int64]Subscription, int, error) {
	subscriptions, count, err := s.Database.QueryList(sortBy, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return subscriptions, count, nil
}
