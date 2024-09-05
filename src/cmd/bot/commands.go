package bot

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
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
	log.Info("Handle /categories command  username=" + c.Sender().Username)

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
	log.Info("Handle /countries command  username=" + c.Sender().Username)

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
	log.Info("Handle /help command  username=" + c.Sender().Username)

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

	return c.Send(output, tele.ModeHTML)
	//return c.Send("Available commands: /help, /subscribe, /unsubscribe, /countries, /datacenters, /plans")
}

func listCommand(k *kimsufi.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /list command username=%s args=%v", c.Sender().Username, args))

		if len(args) < 2 {
			help := "This command list the available plans / servers for a given country and category.\n"
			help += "\n"
			help += "Usage: /list <b>country</b> <b>category</b>\n"
			help += "-  <b>country</b>  You can get the country code by using the /countries command.\n"
			help += "-  <b>category</b> You can get the category by using the /categories command.\n"
			help += "\n"
			help += "Examples:\n"
			help += "  <code>/list fr kimsufi</code>\n"
			help += "  <code>/list de soyoustart</code>\n"
			return c.Send(help, tele.ModeHTML)
		}

		country := strings.ToUpper(args[0])
		if !slices.Contains(kimsufi.AllowedCountries, country) {
			return c.Send(fmt.Sprintf("Invalid country code: <code>%s</code>", country), tele.ModeHTML)
		}

		category := args[1]
		if !slices.Contains(kimsufi.PlanCategories, category) {
			return c.Send(fmt.Sprintf("Invalid category: <code>%s</code>", category), tele.ModeHTML)
		}

		catalog, err := k.ListServers(country)
		if err != nil {
			return fmt.Errorf("failed to list servers: %w", err)
		}

		var output = &bytes.Buffer{}
		w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "planCode\tcategory\tname\tprice")
		fmt.Fprintln(w, "--------\t--------\t----\t-----")

		sort.Slice(catalog.Plans, func(i, j int) bool {
			return catalog.Plans[i].FirstPrice().Price < catalog.Plans[j].FirstPrice().Price
		})

		for _, plan := range catalog.Plans {
			if plan.Blobs.Commercial.Range != category {
				continue
			}

			var price float64
			planPrice := plan.FirstPrice()
			if !reflect.DeepEqual(planPrice, kimsufi.Pricing{}) {
				price = float64(planPrice.Price) / kimsufi.PriceDivider
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\n", plan.PlanCode, category, plan.InvoiceName, price)
		}
		w.Flush()

		return c.Send("<pre>"+output.String()+"</pre>", tele.ModeHTML)
	}
}

func checkCommand(k *kimsufi.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /check command username=%s args=%v", c.Sender().Username, args))

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
func subscribeCommand(k *kimsufi.Service, s *subscription.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /subscribe command username=%s args=%v", c.Sender().Username, args))

		if len(args) < 1 {
			help := "This command subscribes you to a plan / server and notifies you when it becomes available.\n"
			help += "\n"
			help += "Usage: /subscribe <b>planCode</b> <b>datacenter1,datacenter2,...</b>\n"
			help += "-  <b>planCode</b> You can get the plan code by using the /list command.\n"
			help += "-  <b>datacenter1,datacenter2,...</b> Optional. You can specify one or more datacenters to check availability. If not specified, the bot will check all datacenters. Valid datacenters are: <code>" + strings.Join(kimsufi.AllowedDatacenters, "</code>, <code>") + "</code>.\n"
			help += "\n"
			help += "Examples:\n"
			help += "  <code>/subscribe 24ska01</code>\n"
			help += "  <code>/subscribe 24ska01 fra,rbx</code>\n"
			return c.Send(help, tele.ModeHTML)
		}

		planCode := args[0]

		datacenters := []string{}
		if len(args) > 1 {
			datacenters = strings.Split(args[1], ",")
			for _, datacenter := range datacenters {
				if !slices.Contains(kimsufi.AllowedDatacenters, datacenter) {
					return c.Send(fmt.Sprintf("Invalid datacenter: <code>%s</code>", datacenter), tele.ModeHTML)
				}
			}
		}

		availabilities, err := k.GetAvailabilities(datacenters, planCode)
		if err != nil {
			if !kimsufi.IsNotAvailableError(err) {
				log.Errorf("failed to check availability before subscribing: %w", err)
				return c.Send("Failed to check availability before subscribing")
			}
		}
		if len(*availabilities) <= 0 {
			return c.Send(fmt.Sprintf("Invalid plan code: <code>%s</code>", planCode), tele.ModeHTML)
		}

		id, err := s.Subscribe(c.Sender(), planCode, datacenters)
		if err != nil {
			log.Errorf("failed to subscribe: %w", err)
			return fmt.Errorf("Failed to subscribe")
		}

		var datacentersMessage string
		if len(datacenters) > 1 {
			datacentersMessage = "one of the following datacenters"
		} else if len(datacenters) == 1 {
			datacentersMessage = "this datacenter"
		} else {
			datacentersMessage = "any datacenter"
		}

		return c.Send(fmt.Sprintf("You will be notified when plan <code>%s</code> is available in %s <code>%s</code> (subscriptionId: <code>%d</code>)", planCode, datacentersMessage, strings.Join(datacenters, "</code>, <code>"), id), tele.ModeHTML)
	}
}

func unsubscribeCommand(s *subscription.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /unsubscribe command  username=%s args=%v", c.Sender().Username, args))

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
		log.Info("Handle /listsubscriptions command  username=" + c.Sender().Username)

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
