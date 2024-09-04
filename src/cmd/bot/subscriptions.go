package bot

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
)

func startSubscriptionCheck(k *kimsufi.Service, s *subscription.Service, b *tele.Bot) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			checkSubscriptions(k, s, b)
		}
	}()
}

func checkSubscriptions(k *kimsufi.Service, s *subscription.Service, b *tele.Bot) error {
	log.Info("subscriptioncheck: check subscriptions start")
	for _, subscriptions := range s.Subscriptions {
		for id, subscription := range subscriptions {
			availabilities, err := k.GetAvailabilities(subscription.Datacenters, subscription.PlanCode)
			if err != nil {
				log.Errorf("subscriptioncheck: username=%s subscriptionId=%d failed to get availabilities: %v", subscription.User.Username, id, err)
			}

			datacenters := availabilities.GetPlanCodeAvailableDatacenters(subscription.PlanCode)
			if len(datacenters) > 0 {
				_, err = b.Send(subscription.User, "Subscription <code>"+subscription.PlanCode+"</code> is available in <code>"+strings.Join(datacenters, "</code>, <code>")+"</code>", tele.ModeHTML)
				if err != nil {
					log.Errorf("subscriptioncheck: username=%s subscriptionId=%d failed to send message: %v", subscription.User.Username, id, err)
				} else {
					s.Unsubscribe(subscription.User, id)
				}
			}
		}
	}
	log.Info("subscriptioncheck: check subscriptions end")

	return nil
}
