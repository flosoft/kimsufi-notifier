package telegram

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
)

func (b *Bot) unsubscribeWrapper(c tele.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing arguments")
	}
	subscriptionId := args[0]

	log.Infof("unsubscribeWrapper subscriptionId=%s", subscriptionId)

	err := b.unsubscribe(c, subscriptionId)
	if err != nil {
		return fmt.Errorf("error unsubscribing: %w", err)
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) unsubscribe(c tele.Context, id string) error {
	subscriptionId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Send("Invalid subscription ID")
	}

	err = b.subscriptionService.Unsubscribe(c.Sender(), subscriptionId)
	if err != nil {
		if errors.Is(err, subscription.ErrorNotFound) {
			return c.Send("Subscription not found")
		}

		log.Errorf("failed to unsubscribe: %w", err)
		return c.Send("Failed to unsubscribe")
	}

	return c.Send(fmt.Sprintf("You unsubscribed from subscription <code>%d</code>", subscriptionId), tele.ModeHTML)
}

func (b *Bot) unsubscribeCommand(c tele.Context) error {
	return b.listSubscriptions(c, true)
}

func (b *Bot) listSubscriptionsCommand(c tele.Context) error {
	return b.listSubscriptions(c, false)
}

func (b *Bot) listSubscriptions(c tele.Context, showButtons bool) error {
	log.Info("Handle /listsubscriptions command user=" + formatUser(c.Sender()))

	subscriptions, err := b.subscriptionService.ListUser(c.Sender())
	if err != nil {
		if errors.Is(err, subscription.ErrorNotFound) {
			return c.Send("You have no subscriptions")
		}

		log.Errorf("failed to list subscriptions: %w", err)
		return c.Send("Failed to list subscriptions")
	}

	var output = &bytes.Buffer{}
	w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "subscriptionId\tplanCode\tdatacenters\tlast_check")
	fmt.Fprintln(w, "--------------\t--------\t-----------\t----------")

	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	btns := []tele.Btn{}
	for k, v := range subscriptions {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", k, v.PlanCode, strings.Join(v.Datacenters, ", "), v.LastCheck.Format(time.RFC1123))
		s := strconv.FormatInt(k, 10)
		btns = append(btns, m.Data(s, fmt.Sprintf("unsubscribe-%d", k), s))
	}
	w.Flush()
	m.Inline(m.Split(8, btns)...)

	if showButtons {
		return c.Send("<pre>"+output.String()+"</pre>\nSelect a subscription to delete", m, tele.ModeHTML)
	}

	return c.Send("<pre>"+output.String()+"</pre>", tele.ModeHTML)
}
