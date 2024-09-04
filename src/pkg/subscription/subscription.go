package subscription

type Subscription struct {
	PlanCode    string
	Datacenters []string
}

type Service struct {
	Subscriptions map[int]Subscription
}

func NewService() *Service {
	s := &Service{}
	s.Subscriptions = make(map[int]Subscription)

	return s
}

func (s *Service) Subscribe(planCode string, datacenters []string) int {
	s.Subscriptions[len(s.Subscriptions)] = Subscription{PlanCode: planCode, Datacenters: datacenters}

	return len(s.Subscriptions) - 1
}

func (s *Service) Unsubscribe(subscriptionId int) {
	delete(s.Subscriptions, subscriptionId)
}

func (s *Service) List() map[int]Subscription {
	return s.Subscriptions
}
