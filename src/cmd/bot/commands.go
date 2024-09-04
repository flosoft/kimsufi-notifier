package bot

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
)

func categoriesCommand(c tele.Context) error {
	log.Info("Handle /categories command")

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
	log.Info("Handle /countries command")

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
	log.Info("Handle /help command")

	output := "Available commands:\n"
	for _, command := range commands {
		//output += "<code>" + command + "</code>\n"
		output += command + "\n"
	}

	return c.Send(output, tele.ModeHTML)
	//return c.Send("Available commands: /help, /subscribe, /unsubscribe, /countries, /datacenters, /plans")
}

func listCommand(k *kimsufi.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		log.Info("Handle /list command")

		args := c.Args()
		if len(args) < 2 {
			c.Send("Usage: /list <country> <category>")
			return nil
		}

		country := strings.ToUpper(args[0])
		if !slices.Contains(kimsufi.AllowedCountries, country) {
			return c.Send(fmt.Sprintf("Invalid country code: %s", country))
		}

		category := args[1]
		if !slices.Contains(kimsufi.PlanCategories, category) {
			return c.Send(fmt.Sprintf("Invalid category: %s", category))
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
		log.Info("Handle /check command")

		args := c.Args()
		if len(args) < 1 {
			return c.Send("Usage: /check <planCode> [ <datacenter1>,<datacenter2>,... ]")
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
					datacenterMessage = "all datacenters"
				}
				return c.Send(fmt.Sprintf("%s is not available in %s", planCode, datacenterMessage))
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
		log.Info("Handle /subscribe command")

		args := c.Args()
		if len(args) < 1 {
			return c.Send("Usage: /subscribe <planCode> [ <datacenter1>,<datacenter2>,... ]")
		}

		planCode := args[0]

		datacenters := []string{}
		if len(args) > 1 {
			datacenters = strings.Split(args[1], ",")
			for _, datacenter := range datacenters {
				if !slices.Contains(kimsufi.AllowedDatacenters, datacenter) {
					return c.Send(fmt.Sprintf("Invalid datacenter: %s", datacenter))
				}
			}
		}

		availabilities, err := k.GetAvailabilities(datacenters, planCode)
		if err != nil {
			if !kimsufi.IsNotAvailableError(err) {
				return fmt.Errorf("failed to get subscribe: %w", err)
			}
		}
		if len(*availabilities) <= 0 {
			return c.Send(fmt.Sprintf("Invalid plan code: %s", planCode))
		}

		id := s.Subscribe(planCode, datacenters)

		var datacentersMessage string
		if len(datacenters) > 1 {
			datacentersMessage = "one of the following datacenters"
		} else if len(datacenters) == 1 {
			datacentersMessage = "this datacenter"
		} else {
			datacentersMessage = "any datacenter"
		}

		return c.Send(fmt.Sprintf("You will be notified when plan %s is available in %s %s (subscriptionId: %d)", planCode, datacentersMessage, strings.Join(datacenters, ", "), id))
	}
}

func unsubscribeCommand(s *subscription.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		log.Info("Handle /unsubscribe command")

		args := c.Args()
		if len(args) < 1 {
			return c.Send("Usage: /unsubscribe <subscriptionId>")
		}

		subscriptionId, err := strconv.Atoi(args[0])
		if err != nil {
			return c.Send("Invalid subscription ID")
		}

		s.Unsubscribe(subscriptionId)

		return c.Send(fmt.Sprintf("You unsubscribed from subscription %d", subscriptionId))
	}
}

func listSubscriptionsCommand(s *subscription.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		log.Info("Handle /listsubscriptions command")

		subscriptions := s.List()

		var output = &bytes.Buffer{}
		w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "subscriptionId\tplanCode\tdatacenters")
		fmt.Fprintln(w, "--------------\t--------\t-----------")

		for k, v := range subscriptions {
			fmt.Fprintf(w, "%d\t%s\t%s\n", k, v.PlanCode, strings.Join(v.Datacenters, ", "))
		}
		w.Flush()

		return c.Send("<pre>"+output.String()+"</pre>", tele.ModeHTML)
	}
}
