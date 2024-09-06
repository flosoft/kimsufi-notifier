package telegram

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/ovh/go-ovh/ovh"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
)

func (b *Bot) subscribeSelectRegion(c tele.Context) error {
	// Inactive
	log.Info(fmt.Sprintf("subscribeSelectRegion user=%s", formatUser(c.Sender())))

	m := &tele.ReplyMarkup{ResizeKeyboard: true}

	endpoints := []string{}
	for endpoint := range ovh.Endpoints {
		e := strings.Split(endpoint, "-")
		if len(e) >= 2 {
			if e[0] == "ovh" {
				endpoints = append(endpoints, endpoint)
			}
		}
	}
	sort.Strings(endpoints)

	btns := []tele.Btn{}
	for _, endpoint := range endpoints {
		e := strings.Split(endpoint, "-")
		btns = append(btns, m.Data(strings.ToUpper(e[1]), ButtonRegion, endpoint))
	}

	rows := m.Split(8, btns)
	rows = append(rows, m.Row(m.Data("Cancel", ButtonCancel, "cancel")))
	m.Inline(rows...)
	return c.Send("Select a region to list servers from", m)
}

func (b *Bot) subscribeCommand(c tele.Context) error {
	log.Info(fmt.Sprintf("Handle /subscribe command user=%s", formatUser(c.Sender())))

	m := &tele.ReplyMarkup{ResizeKeyboard: true}

	btns := []tele.Btn{}
	for _, region := range kimsufi.AllowedRegions {
		for _, country := range region.Countries {
			btns = append(btns, m.Data(region.Name+": "+strings.ToUpper(country), ButtonCountry, region.Endpoint, country))
		}
	}

	rows := m.Split(2, btns)
	rows = append(rows, m.Row(m.Data("Cancel", ButtonCancel, "cancel")))
	m.Inline(rows...)
	err := c.Send("Select a country or region to list servers from", m)
	if err != nil {
		log.Errorf("subscribeSelectCountry failed to send message: %v", err)
		return err
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) listSelectCategory(c tele.Context, args []string) error {
	if len(args) < 2 {
		log.Errorf("listSelectCategory missing arguments args=%d", len(args))
		return c.Edit("Failed to fetch categories")
	}

	region := args[0]
	country := args[1]

	log.Info(fmt.Sprintf("listSelectCategory user=%s country=%s", formatUser(c.Sender()), country))

	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	btns := []tele.Btn{}
	for _, category := range kimsufi.PlanCategories {
		if category != "" {
			btns = append(btns, m.Data(category, ButtonCategory, region, country, category))
		}
	}

	rows := m.Split(8, btns)
	rows = append(rows, m.Row(m.Data("Cancel", ButtonCancel, "cancel")))
	m.Inline(rows...)
	err := c.Edit("Select a server category", m)
	if err != nil {
		log.Errorf("listSelectCategory failed to send message: %v", err)
		return err
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) listWrapper(c tele.Context, args []string) error {
	if len(args) < 2 {
		log.Errorf("listWrapper missing arguments args=%d", len(args))
		return c.Edit("Failed to fetch servers")
	}

	region := args[0]
	country := args[1]
	category := args[2]

	log.Infof("listWrapper user=%s region=%s country=%s category=%s", formatUser(c.Sender()), region, country, category)

	err := b.list(c, region, country, category)
	if err != nil {
		log.Errorf("listWrapper failed to list servers: %v", err)
		return c.Edit("Failed to fetch servers")
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) list(c tele.Context, region, country, category string) error {
	//if !slices.Contains(kimsufi.AllowedCountries, country) {
	//	return c.Edit(fmt.Sprintf("Invalid country code: <code>%s</code>", country), tele.ModeHTML)
	//}

	if !slices.Contains(kimsufi.PlanCategories, category) {
		return c.Edit(fmt.Sprintf("Invalid category: <code>%s</code>", category), tele.ModeHTML)
	}

	catalog, err := b.kimsufiService.Endpoint(region).ListServers(country)
	if err != nil {
		log.Errorf("list: failed to list servers: %v", err)
		return c.Edit("Failed to fetch servers")
	}

	var output = &bytes.Buffer{}
	w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "planCode\tcategory\tname\tprice")
	fmt.Fprintln(w, "--------\t--------\t----\t-----")

	sort.Slice(catalog.Plans, func(i, j int) bool {
		return catalog.Plans[i].FirstPrice().Price < catalog.Plans[j].FirstPrice().Price
	})

	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	btns := []tele.Btn{}

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

		btns = append(btns, m.Data(plan.PlanCode, ButtonPlanCode, region, country, category, plan.PlanCode))
	}
	w.Flush()
	rows := m.Split(4, btns)
	rows = append(rows, m.Row(m.Data("Cancel", ButtonCancel, "cancel")))
	m.Inline(rows...)

	if len(catalog.Plans) == 0 {
		return c.Edit("No servers found")
	}

	return c.Edit("<pre>"+output.String()+"</pre>Select which server you want to be notified about", m, tele.ModeHTML)
}

func (b *Bot) listCommand(k *kimsufi.Service) func(tele.Context) error {
	return func(c tele.Context) error {
		args := c.Args()

		log.Info(fmt.Sprintf("Handle /list command user=%s args=%v", formatUser(c.Sender()), args))

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
		region := args[0]
		country := strings.ToUpper(args[1])
		category := args[2]

		return b.list(c, region, country, category)
	}
}

func parseUniqueData(input []string) ([]string, []string) {
	unique := strings.TrimSpace(input[0])
	//data := strings.TrimSpace(input[1])
	uniqueValues := strings.Split(unique, "-")
	return input[1:], uniqueValues
}
