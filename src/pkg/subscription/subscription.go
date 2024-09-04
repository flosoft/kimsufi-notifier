package subscription

import (
	tele "gopkg.in/telebot.v3"
)

type Subscription struct {
	PlanCode    string
	Datacenters []string
}

type Service struct {
	Subscriptions map[*tele.User]map[int]Subscription
}

func NewService() *Service {
	s := &Service{}
	s.Subscriptions = make(map[*tele.User]map[int]Subscription)

	return s
}

func (s *Service) Subscribe(telegramUser *tele.User, planCode string, datacenters []string) int {
	if _, ok := s.Subscriptions[telegramUser]; !ok {
		s.Subscriptions[telegramUser] = make(map[int]Subscription)
	}

	s.Subscriptions[telegramUser][len(s.Subscriptions[telegramUser])] = Subscription{
		PlanCode:    planCode,
		Datacenters: datacenters,
	}

	return len(s.Subscriptions[telegramUser]) - 1
}

func (s *Service) Unsubscribe(telegramUser *tele.User, subscriptionId int) {
	if _, ok := s.Subscriptions[telegramUser]; !ok {
		return
	}

	delete(s.Subscriptions[telegramUser], subscriptionId)
}

func (s *Service) List(telegramUser *tele.User) map[int]Subscription {
	if _, ok := s.Subscriptions[telegramUser]; !ok {
		return nil
	}

	return s.Subscriptions[telegramUser]
}
