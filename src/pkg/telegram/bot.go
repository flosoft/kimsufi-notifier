package telegram

import (
	"fmt"
	"os"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
)

func NewBot() (*tele.Bot, error) {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	b.Handle("/hello", func(c tele.Context) error {
		username := c.Sender().Username

		return c.Send("Hello! @" + username)
	})

	b.Handle("/help", func(c tele.Context) error {
		return c.Send("Available commands: /help, /subscribe, /unsubscribe, /countries, /datacenters, /plans")
	})

	b.Handle("/subscribe", func(c tele.Context) error {
		username := c.Sender().Username
		fmt.Printf("payload: %s\n", c.Message().Payload)
		var planCode string
		var datacentersString string
		n, err := fmt.Sscan(c.Message().Payload, &planCode, &datacentersString)
		if err != nil {
			fmt.Println("Invalid command")
			return err
		}
		var datacenters = strings.Split(strings.Trim(datacentersString, ","), ",")
		fmt.Printf("scanned %d variables: %s, %v\n", n, planCode, datacenters)

		var datacentersMessage string
		if len(datacenters) > 0 {
			datacentersMessage = "one of the following datacenters"
		} else {
			datacentersMessage = "this datacenter"
		}
		subscriptionId := 1

		return c.Send(fmt.Sprintf("@%s you will be notified when plan %s is available in %s %s (subscriptionId: %d)", username, planCode, datacentersMessage, strings.Join(datacenters, ", "), subscriptionId))
	})

	b.Handle("/unsubscribe", func(c tele.Context) error {
		username := c.Sender().Username
		fmt.Printf("payload: %s\n", c.Message().Payload)
		var subscriptionId int
		n, err := fmt.Sscan(c.Message().Payload, &subscriptionId)
		if err != nil {
			fmt.Println("Invalid command")
			return err
		}

		fmt.Printf("scanned %d variables: %d\n", n, subscriptionId)

		return c.Send(fmt.Sprintf("@%s you unsubscribed from subscription %d", username, subscriptionId))
	})

	b.Handle("/countries", func(c tele.Context) error {
		return c.Send(fmt.Sprintf("Allowed countries: %s", strings.Join(kimsufi.AllowedCountries, ", ")))
	})

	b.Handle("/datacenters", func(c tele.Context) error {
		return c.Send(fmt.Sprintf("Allowed datacenters: %s", strings.Join(kimsufi.AllowedDatacenters, ", ")))
	})

	return b, nil
}
