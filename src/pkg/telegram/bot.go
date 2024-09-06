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
	kimsufiService      *kimsufi.MultiService
	subscriptionService *subscription.Service
}

func NewBot(k *kimsufi.MultiService, s *subscription.Service) (*Bot, error) {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	b.Handle("/hello", func(c tele.Context) error {
		log.Info(fmt.Sprintf("Handle /hello command user=%s", formatUser(c.Sender())))

		username := c.Sender().Username

		return c.Send("Test notification! @" + username)
	})

	bot := &Bot{
		Bot:                 b,
		kimsufiService:      k,
		subscriptionService: s,
	}

	b.Handle(tele.OnText, helpCommand)
	b.Handle(commands["help"].command, helpCommand)
	b.Handle(commands["subscribe"].command, bot.subscribeCommand)
	b.Handle(commands["unsubscribe"].command, bot.unsubscribeCommand)
	b.Handle(commands["listsubscriptions"].command, bot.listSubscriptionsCommand)

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		callback := c.Callback()

		//log.WithField("rawData", callback.Data).WithField("rawUnique", callback.Unique).Trace("Callback rawData")
		garbage := strings.Split(callback.Data, "|")
		data, values := parseUniqueData(garbage)
		log.WithField("data", strings.Join(data, "|")).WithField("unique", strings.Join(values, "-")).Trace("Callback parsedData")

		switch values[0] {
		//case ButtonRegion:
		//	return bot.subscribeSelectCountry(c, data)
		case ButtonCountry:
			return bot.listSelectCategory(c, data)
		case ButtonCategory:
			return bot.listWrapper(c, data)
		case ButtonPlanCode:
			return bot.subscribeSelectDatacenters(c, data)
		case ButtonDatacenter:
			return bot.subscribeWrapper(c, data)
		case ButtontUnsubscribe:
			return bot.unsubscribeWrapper(c, data)
		case ButtonCancel:
			return bot.cancelWrapper(c)
		}

		return c.Respond(&tele.CallbackResponse{})
	})

	return bot, nil
}

func formatUser(user *tele.User) string {
	return fmt.Sprintf("%s %s (@%s)", user.FirstName, user.LastName, user.Username)
}
