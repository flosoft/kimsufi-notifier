package bot

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
)

var (
	subscriptionCheckInterval = 5 * time.Minute
	subscriptionCheckLimit    = 100
	subscriptionCheckOffset   = 0
)

func startSubscriptionCheck(m *kimsufi.MultiService, s *subscription.Service, b *tele.Bot) {
	ticker := time.NewTicker(subscriptionCheckInterval)

	go func() {
		for range ticker.C {
			checkSubscriptions(m, s, b)
		}
	}()
}

func checkSubscriptions(m *kimsufi.MultiService, s *subscription.Service, b *tele.Bot) error {
	currentOffset := subscriptionCheckOffset
	for {
		subscriptions, _, err := s.ListPaginate("user_id", subscriptionCheckLimit, currentOffset)
		if err != nil {
			return err
		}

		if len(subscriptions) == 0 {
			break
		}

		for _, subscription := range subscriptions {
			log.Infof("subscriptioncheck: subscriptionId=%d check", subscription.ID)
			availabilities, err := m.Endpoint(subscription.Region).GetAvailabilities(subscription.Datacenters, subscription.PlanCode)
			if err != nil {
				log.Errorf("subscriptioncheck: subscriptionId=%d failed to get availabilities: %v", subscription.ID, err)
			}

			datacenters := availabilities.GetPlanCodeAvailableDatacenters(subscription.PlanCode)
			if len(datacenters) > 0 {
				for _, user := range subscription.Users {
					_, err = b.Send(user, fmt.Sprintf("@%s plan <code>%s</code> is available in <code>%s</code>", user.Username, subscription.PlanCode, strings.Join(datacenters, "</code>, <code>")), tele.ModeHTML)
					if err != nil {
						log.Errorf("subscriptioncheck: subscriptionId=%d username=%s failed to send message: %v", subscription.ID, user.Username, err)
					} else {
						subscription.Notifications++
						s.Unsubscribe(user, subscription.ID)
					}
				}
				s.Database.UpdateNotifications(subscription.ID, subscription.Notifications)
			}
			s.Database.UpdateLastCheck(subscription.ID)
		}

		currentOffset = currentOffset + subscriptionCheckLimit
	}

	return nil
}
