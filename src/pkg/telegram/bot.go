package telegram

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
)

type Bot struct {
	*tele.Bot
	kimsufiService      *kimsufi.Service
	subscriptionService *subscription.Service
}

func NewBot(k *kimsufi.Service, s *subscription.Service) (*Bot, error) {
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

	bot := &Bot{
		Bot:                 b,
		kimsufiService:      k,
		subscriptionService: s,
	}

	b.Handle(commands["subscribe"].command, bot.listSelectCountry)
	b.Handle(commands["unsubscribe"].command, bot.unsubscribeCommand)
	b.Handle(commands["listsubscriptions"].command, bot.listSubscriptionsCommand)

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		callback := c.Callback()
		log.WithField("rawData", callback.Data).WithField("rawUnique", callback.Unique).Info("Callback rawData")
		garbage := strings.Split(callback.Data, "|")
		data, values := parseUniqueData(garbage)
		log.WithField("data", strings.Join(data, "|")).WithField("unique", strings.Join(values, "-")).Info("Callback parsedData")

		switch values[0] {
		case "listcountry":
			return bot.listSelectCategory(c, data)
		case "listcategory":
			return bot.listWrapper(c, data)
		case "listplancode":
			return bot.subscribeSelectDatacenters(c, data)
		case "subscribedatacenters":
			return bot.subscribeWrapper(c, data)
		case "unsubscribe":
			return bot.unsubscribeWrapper(c, data)
		}

		return c.Respond(&tele.CallbackResponse{})
	})

	return bot, nil
}

func formatUser(user *tele.User) string {
	return fmt.Sprintf("%s %s (@%s)", user.FirstName, user.LastName, user.Username)
}
