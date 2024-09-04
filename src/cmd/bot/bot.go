package bot

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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
	commands = map[string]string{
		"help":              "/help",
		"categories":        "/categories",
		"countries":         "/countries",
		"list":              "/list",
		"check":             "/check",
		"subscribe":         "/subscribe",
		"unsubscribe":       "/unsubscribe",
		"listsubscriptions": "/listsubscriptions",
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

	telegramBot.Handle(commands["help"], helpCommand)
	telegramBot.Handle(commands["categories"], categoriesCommand)
	telegramBot.Handle(commands["countries"], countriesCommand)
	telegramBot.Handle(commands["list"], listCommand(k))
	telegramBot.Handle(commands["check"], checkCommand(k))
	telegramBot.Handle(commands["subscribe"], subscribeCommand(k, s))
	telegramBot.Handle(commands["unsubscribe"], unsubscribeCommand(s))
	telegramBot.Handle(commands["listsubscriptions"], listSubscriptionsCommand(s))

	startSubscriptionCheck(k, s, telegramBot)

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
