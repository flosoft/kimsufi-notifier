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
		log.Errorf("unsubscribeWrapper missing arguments args=%d", len(args))
		return c.Edit("Failed to unsubscribe")
	}
	subscriptionId := args[0]

	log.Infof("unsubscribeWrapper user=%s subscriptionId=%s", formatUser(c.Sender()), subscriptionId)

	err := b.unsubscribe(c, subscriptionId)
	if err != nil {
		log.Errorf("unsubscribeWrapper error unsubscribing: %v", err)
		return c.Edit("Failed to unsubscribe")
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) unsubscribe(c tele.Context, id string) error {
	if id == "all" {
		err := b.subscriptionService.UnsubscribeAll(c.Sender())
		if err != nil {
			if errors.Is(err, subscription.ErrorNotFound) {
				return c.Edit("No subscriptions")
			}

			log.Errorf("unsubscribe failed to unsubscribe: %v", err)
			return c.Edit("Failed to unsubscribe")
		}

		return c.Edit("You unsubscribed from all notification", tele.ModeHTML)
	}

	subscriptionId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Errorf("unsubscribe invalid subscription ID: %v", err)
		return c.Edit("Invalid subscription ID")
	}

	err = b.subscriptionService.Unsubscribe(c.Sender(), subscriptionId)
	if err != nil {
		if errors.Is(err, subscription.ErrorNotFound) {
			return c.Edit("Subscription not found")
		}

		log.Errorf("unsubscribe failed to unsubscribe: %v", err)
		return c.Edit("Failed to unsubscribe")
	}

	return c.Edit(fmt.Sprintf("You unsubscribed from notification <code>%d</code>", subscriptionId), tele.ModeHTML)

}

func (b *Bot) unsubscribeCommand(c tele.Context) error {
	log.Info("Handle /unsubscribe command user=" + formatUser(c.Sender()))

	return b.listSubscriptions(c, true)
}

func (b *Bot) listSubscriptionsCommand(c tele.Context) error {
	log.Info("Handle /listsubscriptions command user=" + formatUser(c.Sender()))

	return b.listSubscriptions(c, false)
}

func (b *Bot) listSubscriptions(c tele.Context, showButtons bool) error {
	subscriptions, err := b.subscriptionService.ListUser(c.Sender())
	if err != nil {
		if errors.Is(err, subscription.ErrorNotFound) {
			return c.Send("You have no subscriptions")
		}

		log.Errorf("listSubscriptions failed to list subscriptions: %v", err)
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
		btns = append(btns, m.Data(s, ButtontUnsubscribe, s))
	}
	w.Flush()
	rows := m.Split(8, btns)
	rows = append(rows, m.Row(m.Data("Unsubscribe from all", ButtontUnsubscribe, "all")))
	rows = append(rows, m.Row(m.Data("Cancel", ButtonCancel, "cancel")))
	m.Inline(rows...)

	if len(subscriptions) == 0 {
		return c.Send("You have no subscriptions")
	}

	if showButtons {
		return c.Send("<pre>"+output.String()+"</pre>\nSelect a subscription to delete", m, tele.ModeHTML)
	}

	return c.Send("<pre>"+output.String()+"</pre>", tele.ModeHTML)
}
