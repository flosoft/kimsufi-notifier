package bot

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v3"

	"github.com/TheoBrigitte/kimsufi-notifier/cmd/flag"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/subscription"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/telegram"
)

var (
	Cmd = &cobra.Command{
		Use:   "bot",
		Short: "Run Telegram bot",
		RunE:  runner,
	}

	planCode      string
	ovhSubsidiary string
)

func init() {
}

var (
	commands = map[string]struct {
		command string
		help    string
	}{
		"help": {
			"/help",
			"Show this help",
		},
		"categories": {
			"/categories",
			"List available categories",
		},
		"countries": {
			"/countries",
			"List available countries",
		},
		"list": {
			"/list",
			"List available plans / servers",
		},
		"check": {
			"/check",
			"Check availability of a plan",
		},
		"subscribe": {
			"/subscribe",
			"Get notified when a server becomes available",
		},
		"unsubscribe": {
			"/unsubscribe",
			"Unsubscribe from a notification",
		},
		"listsubscriptions": {
			"/listsubscriptions",
			"List active subscriptions",
		},
	}
)

func runner(cmd *cobra.Command, args []string) error {
	d := kimsufi.Config{
		URL:    kimsufi.GetOVHEndpoint(cmd.Flag(flag.OVHAPIEndpointFlagName).Value.String()),
		Logger: log.StandardLogger(),
	}
	k, err := kimsufi.NewService(d)
	if err != nil {
		return fmt.Errorf("failed to initialize kimsufi service: %w", err)
	}

	s := subscription.NewService()

	telegramBot, err := telegram.NewBot()
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	telegramBot.Handle(commands["help"].command, helpCommand)
	telegramBot.Handle(commands["categories"].command, categoriesCommand)
	telegramBot.Handle(commands["countries"].command, countriesCommand)
	telegramBot.Handle(commands["list"].command, listCommand(k))
	telegramBot.Handle(commands["check"].command, checkCommand(k))
	telegramBot.Handle(commands["subscribe"].command, subscribeCommand(k, s))
	telegramBot.Handle(commands["unsubscribe"].command, unsubscribeCommand(s))
	telegramBot.Handle(commands["listsubscriptions"].command, listSubscriptionsCommand(s))
	telegramBot.Handle(tele.OnText, helpCommand)

	startSubscriptionCheck(k, s, telegramBot)

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
