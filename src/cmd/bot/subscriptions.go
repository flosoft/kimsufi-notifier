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
	subscriptionCheckInterval = 5 * time.Second
)

func startSubscriptionCheck(k *kimsufi.Service, s *subscription.Service, b *tele.Bot) {
	ticker := time.NewTicker(subscriptionCheckInterval)

	go func() {
		for range ticker.C {
			checkSubscriptions(k, s, b)
		}
	}()
}

func checkSubscriptions(k *kimsufi.Service, s *subscription.Service, b *tele.Bot) error {
	var limit = 100
	var offset = 0
	for {
		subscriptions, _, err := s.ListPaginate("user_id", limit, offset)
		if err != nil {
			return err
		}

		if len(subscriptions) == 0 {
			break
		}

		for _, subscriptions := range subscriptions {
			for id, subscription := range subscriptions {
				log.Infof("subscriptioncheck: username=%s subscriptionId=%d check", subscription.User.Username, id)
				availabilities, err := k.GetAvailabilities(subscription.Datacenters, subscription.PlanCode)
				if err != nil {
					log.Errorf("subscriptioncheck: username=%s subscriptionId=%d failed to get availabilities: %v", subscription.User.Username, id, err)
				}

				datacenters := availabilities.GetPlanCodeAvailableDatacenters(subscription.PlanCode)
				if len(datacenters) > 0 {
					_, err = b.Send(subscription.User, fmt.Sprintf("@%s plan <code>%s</code> is available in <code>%s</code>", subscription.User.Username, subscription.PlanCode, strings.Join(datacenters, "</code>, <code>")), tele.ModeHTML)
					if err != nil {
						log.Errorf("subscriptioncheck: username=%s subscriptionId=%d failed to send message: %v", subscription.User.Username, id, err)
					} else {
						s.Unsubscribe(subscription.User, id)
					}
				}
			}
		}
	}

	return nil
}
