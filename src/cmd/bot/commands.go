package bot

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

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
)

func categoriesCommand(c tele.Context) error {
	log.Info("Handle /categories command user=" + formatUser(c.Sender()))

	categories := kimsufi.PlanCategories

	output := "Available categories:\n"
	for _, category := range categories {
		if category != "" {
			output += "<code>" + category + "</code>\n"
		}
	}

	return c.Send(output, tele.ModeHTML)
}

func countriesCommand(c tele.Context) error {
	log.Info("Handle /countries command user=" + formatUser(c.Sender()))

	countries := kimsufi.AllowedCountries

	output := "Allowed countries:\n"
	for _, country := range countries {
		if country != "" {
			output += "<code>" + country + "</code>\n"
		}
	}

	return c.Send(output, tele.ModeHTML)
}

func helpCommand(c tele.Context) error {
	log.Info("Handle /help command user=" + formatUser(c.Sender()))

	output := "Hello,\n"
	output += "This bot can help you to monitor the availability of Kimsufi servers.\n"
	output += "\n"
	output += "You can subscribe to a plan and get notified when it becomes available with /subscribe command.\n"
	output += "You can also list available servers and check their availability with /list command.\n"
	output += "\n"
	output += "You can use the following commands:\n"

	for _, command := range commands {
		output += command.command + "  " + command.help + "\n"
	}

	output += "\n"
	output += "Ask for support or report issues in our Telegram group: https://t.me/+xPnf7KSGEoA1Nzcy\n"

	return c.Send(output, tele.ModeHTML)
	//return c.Send("Available commands: /help, /subscribe, /unsubscribe, /countries, /datacenters, /plans")
}

func checkCommand(k *kimsufi.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /check command user=%s args=%v", formatUser(c.Sender()), args))

		if len(args) < 1 {
			help := "This command checks the availability of a given plan / server in one or more datacenters.\n"
			help += "\n"
			help += "Usage: /check <b>planCode</b> <b>datacenter1,datacenter2,...</b>\n"
			help += "-  <b>planCode</b> You can get the plan code by using the /list command.\n"
			help += "-  <b>datacenter1,datacenter2,...</b> Optional. You can specify one or more datacenters to check availability. If not specified, the bot will check all datacenters. Valid datacenters are: <code>" + strings.Join(kimsufi.AllowedDatacenters, "</code>, <code>") + "</code>.\n"
			help += "\n"
			help += "Examples:\n"
			help += "  <code>/check 24ska01</code>\n"
			help += "  <code>/check 24ska01 fra,rbx</code>\n"
			return c.Send(help, tele.ModeHTML)
		}

		planCode := args[0]
		datacenters := []string{}
		if len(args) > 1 {
			datacenters = strings.Split(args[1], ",")
		}

		availabilities, err := k.GetAvailabilities(datacenters, planCode)
		if err != nil {
			if kimsufi.IsNotAvailableError(err) {
				datacenterMessage := ""
				if len(datacenters) > 0 {
					datacenterMessage = strings.Join(datacenters, ", ")
				} else {
					datacenterMessage = "any datacenters"
				}
				return c.Send(fmt.Sprintf("<code>%s</code> is not available in %s", planCode, datacenterMessage), tele.ModeHTML)
			} else {
				return fmt.Errorf("failed to get availabilities: %w", err)
			}
		}

		formatter := kimsufi.DatacenterFormatter(kimsufi.IsDatacenterAvailable, kimsufi.DatacenterKey)
		result := availabilities.Format(kimsufi.PlanCode, formatter)

		var output = &bytes.Buffer{}
		w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "planCode\tstatus\tdatacenters")
		fmt.Fprintln(w, "--------\t------\t-----------")

		for k, v := range result {
			status := "available"
			if len(v) == 0 {
				status = "unavailable"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", k, status, strings.Join(v, ", "))
		}
		w.Flush()

		return c.Send("<pre>"+output.String()+"</pre>", tele.ModeHTML)
	}
}

func unsubscribeCommand(s *subscription.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /unsubscribe command user=%s args=%v", formatUser(c.Sender()), args))

		if len(args) < 1 {
			help := "This command removes a subscription you have created with /subscribe command.\n"
			help += "\n"
			help += "Usage: /unsubscribe <b>subscriptionId</b>\n"
			help += "-  <b>subscriptionId</b> You can get the subscription ID by using the /listsubscriptions command.\n"
			help += "\n"
			help += "Examples:\n"
			help += "  <code>/unsubscribe 1</code>\n"
			return c.Send(help, tele.ModeHTML)
		}

		subscriptionId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return c.Send("Invalid subscription ID")
		}

		err = s.Unsubscribe(c.Sender(), subscriptionId)
		if err != nil {
			if errors.Is(err, subscription.ErrorNotFound) {
				return c.Send("Subscription not found")
			}

			log.Errorf("failed to unsubscribe: %w", err)
			return c.Send("Failed to unsubscribe")
		}

		return c.Send(fmt.Sprintf("You unsubscribed from subscription <code>%d</code>", subscriptionId), tele.ModeHTML)
	}
}

func listSubscriptionsCommand(s *subscription.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		log.Info("Handle /listsubscriptions command user=" + formatUser(c.Sender()))

		subscriptions, err := s.ListUser(c.Sender())
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

		for k, v := range subscriptions {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", k, v.PlanCode, strings.Join(v.Datacenters, ", "), v.LastCheck.Format(time.RFC1123))
		}
		w.Flush()

		return c.Send("<pre>"+output.String()+"</pre>", tele.ModeHTML)
	}
}

func formatUser(user *tele.User) string {
	return fmt.Sprintf("%s %s (@%s)", user.FirstName, user.LastName, user.Username)
}
