package subscription

import (
	tele "gopkg.in/telebot.v3"
)

type Subscription struct {
	PlanCode    string
	Datacenters []string
	User        *tele.User
}

type Service struct {
	Subscriptions map[int64]map[int]Subscription
}

func NewService() *Service {
	s := &Service{}
	s.Subscriptions = make(map[int64]map[int]Subscription)

	return s
}

func (s *Service) Subscribe(telegramUser *tele.User, planCode string, datacenters []string) int {
	if _, ok := s.Subscriptions[telegramUser.ID]; !ok {
		s.Subscriptions[telegramUser.ID] = make(map[int]Subscription)
	}

	s.Subscriptions[telegramUser.ID][len(s.Subscriptions[telegramUser.ID])] = Subscription{
		PlanCode:    planCode,
		Datacenters: datacenters,
		User:        telegramUser,
	}

	return len(s.Subscriptions[telegramUser.ID]) - 1
}

func (s *Service) Unsubscribe(telegramUser *tele.User, subscriptionId int) {
	if _, ok := s.Subscriptions[telegramUser.ID]; !ok {
		return
	}

	delete(s.Subscriptions[telegramUser.ID], subscriptionId)
}

func (s *Service) List(telegramUser *tele.User) map[int]Subscription {
	if _, ok := s.Subscriptions[telegramUser.ID]; !ok {
		return nil
	}

	return s.Subscriptions[telegramUser.ID]
}
