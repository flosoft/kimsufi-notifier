package telegram

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
)

func (b *Bot) subscribeSelectDatacenters(c tele.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("missing arguments")
	}
	country := args[0]
	category := args[1]
	planCode := args[2]

	log.Infof("subscribeSelectDatacenters country=%s category=%s planCode=%s", country, category, planCode)

	catalog, err := b.kimsufiService.ListServers(country)
	if err != nil {
		return fmt.Errorf("Failed to list servers: %w", err)
	}

	plan := catalog.GetPlan(planCode)
	if plan == nil {
		return c.Send("Server not found")
	}

	datacenters := plan.GetDatacenters()
	log.Infof("planCode=%s config=%d datacenters=%v", plan.PlanCode, len(plan.Configurations), datacenters)
	if len(datacenters) <= 0 {
		err := b.subscribe(c, planCode, "")
		if err != nil {
			return fmt.Errorf("error subscribing: %w", err)
		}
		return c.Respond(&tele.CallbackResponse{})
	}

	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	btns := []tele.Btn{}

	for _, datacenter := range datacenters {
		btns = append(btns, m.Data(datacenter, "subscribedatacenters-"+datacenter, country, category, planCode, datacenter))
	}

	rows := m.Split(8, btns)
	rows = append(rows, m.Row(m.Data("any datacenter", "subscribedatacenters-any", country, category, planCode, "any")))
	m.Inline(rows...)
	err = c.Send("Select a datacenter", m)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) subscribeWrapper(c tele.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("subscribeWrapper: missing arguments %d", len(args))
	}
	country := args[0]
	category := args[1]
	planCode := args[2]
	datacenters := args[3]

	log.Infof("subscribeWrapper country=%s category=%s planCode=%s datacenters=%s", country, category, planCode, datacenters)

	d := ""
	if datacenters != "any" {
		d = datacenters
	}

	err := b.subscribe(c, planCode, d)
	if err != nil {
		return fmt.Errorf("error subscribing: %w", err)
	}

	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) subscribe(c tele.Context, planCode, datacentersString string) error {
	log.Infof("subscribe planCode=%s datacenters=%s", planCode, datacentersString)

	datacenters := []string{}
	if len(datacentersString) > 0 {
		datacenters = strings.Split(datacentersString, ",")
	}

	_, err := b.kimsufiService.GetAvailabilities(datacenters, planCode)
	if err != nil {
		if !kimsufi.IsNotAvailableError(err) {
			log.Errorf("failed to check availability before subscribing: %w", err)
			return c.Send("Failed to check availability before subscribing")
		}
	}
	// This code might not be required, empty availability used to mean invalid plan.
	//if len(*availabilities) <= 0 {
	//	return c.Send(fmt.Sprintf("Invalid plan code: <code>%s</code>", planCode), tele.ModeHTML)
	//}

	id, err := b.subscriptionService.Subscribe(c.Sender(), planCode, datacenters)
	if err != nil {
		log.Errorf("failed to subscribe: %w", err)
		return fmt.Errorf("Failed to subscribe")
	}

	var datacentersMessage string
	if len(datacenters) > 1 {
		datacentersMessage = fmt.Sprintf("one of the following datacenters <code>%s</code>", strings.Join(datacenters, "</code>, <code>"))
	} else if len(datacenters) == 1 {
		datacentersMessage = fmt.Sprintf("<code>%s</code> datacenter", datacenters[0])
	} else {
		datacentersMessage = "any datacenter"
	}

	return c.Send(fmt.Sprintf("You will be notified when plan <code>%s</code> is available in %s (subscriptionId: <code>%d</code>)", planCode, datacentersMessage, id), tele.ModeHTML)
}

func (b *Bot) subscribeCommand(c tele.Context) error {
	args := c.Args()

	log.Info(fmt.Sprintf("Handle /subscribe command user=%s args=%v", formatUser(c.Sender()), args))

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
	datacenters := ""
	if len(args) > 1 {
		datacenters = args[1]
	}

	return b.subscribe(c, planCode, datacenters)

}
