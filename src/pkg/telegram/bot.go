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
