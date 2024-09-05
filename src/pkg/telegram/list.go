package telegram

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
)

func (b *Bot) listSelectCountry(c tele.Context) error {
	//args := c.Args()
	//log.Info(fmt.Sprintf("Handle list command user=%s args=%v", formatUser(c.Sender()), args))

	log.Infof("listSelectCountry")

	m := &tele.ReplyMarkup{ResizeKeyboard: true}

	btns := []tele.Btn{}
	for _, country := range kimsufi.AllowedCountries {
		btns = append(btns, m.Data(country, "listcountry-"+country, country))
	}

	m.Inline(m.Split(8, btns)...)
	return c.Send("Select a country to list servers from", m)
}

func (b *Bot) listSelectCategory(c tele.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing arguments")
	}
	country := args[0]

	log.Infof("listSelectCategory country=%s", country)
	//data, _ := parseUniqueData(args[0:2])
	//country := data

	m := &tele.ReplyMarkup{ResizeKeyboard: true}

	btns := []tele.Btn{}
	for _, category := range kimsufi.PlanCategories {
		if category != "" {
			btns = append(btns, m.Data(category, "listcategory-"+category, country, category))
		}
	}

	m.Inline(m.Split(8, btns)...)
	err := c.Send("Select a server category", m)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) listWrapper(c tele.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing arguments")
	}
	country := args[0]
	category := args[1]

	log.Infof("listWrapper country=%s category=%s", country, category)

	err := b.list(c, country, category)
	if err != nil {
		return fmt.Errorf("error listing servers: %w", err)
	}
	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) list(c tele.Context, country, category string) error {
	if !slices.Contains(kimsufi.AllowedCountries, country) {
		return c.Send(fmt.Sprintf("Invalid country code: <code>%s</code>", country), tele.ModeHTML)
	}

	if !slices.Contains(kimsufi.PlanCategories, category) {
		return c.Send(fmt.Sprintf("Invalid category: <code>%s</code>", category), tele.ModeHTML)
	}

	catalog, err := b.kimsufiService.ListServers(country)
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	var output = &bytes.Buffer{}
	w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "category\tname\tprice")
	fmt.Fprintln(w, "--------\t----\t-----")

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

		fmt.Fprintf(w, "%s\t%s\t%.2f\n", category, plan.InvoiceName, price)

		shortName := strings.Split(plan.InvoiceName, " | ")[0]

		btns = append(btns, m.Data(shortName, "listplancode-"+plan.PlanCode, country, category, plan.PlanCode))
	}
	w.Flush()

	m.Inline(m.Split(4, btns)...)

	return c.Send("<pre>"+output.String()+"</pre>Select which server you want to be notified about", m, tele.ModeHTML)
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
		country := strings.ToUpper(args[0])
		category := args[1]

		return b.list(c, country, category)
	}
}

func parseUniqueData(input []string) ([]string, []string) {
	unique := strings.TrimSpace(input[0])
	//data := strings.TrimSpace(input[1])
	uniqueValues := strings.Split(unique, "-")
	return input[1:], uniqueValues
}
