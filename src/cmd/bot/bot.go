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

	databaseFilename string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&databaseFilename, "database-filename", "d", "kimsufi-notifier.sqlite3", "filename of the SQLite database")
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
		"hello": {
			"/hello",
			"Send a test notification",
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

	s, err := subscription.NewService(databaseFilename)
	if err != nil {
		return fmt.Errorf("failed to initialize subscription service: %w", err)
	}

	telegramBot, err := telegram.NewBot(k, s)
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	telegramBot.Handle(commands["help"].command, helpCommand)
	telegramBot.Handle(tele.OnText, helpCommand)

	startSubscriptionCheck(k, s, telegramBot.Bot)

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
