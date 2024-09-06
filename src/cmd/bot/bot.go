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

	databaseFilename string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&databaseFilename, "database-filename", "d", "kimsufi-notifier.sqlite3", "filename of the SQLite database")
}

func runner(cmd *cobra.Command, args []string) error {
	d := kimsufi.Config{
		URL:    kimsufi.GetOVHEndpoint(cmd.Flag(flag.OVHAPIEndpointFlagName).Value.String()),
		Logger: log.StandardLogger(),
	}
	m, err := kimsufi.NewService(d)
	if err != nil {
		return fmt.Errorf("failed to initialize kimsufi service: %w", err)
	}

	s, err := subscription.NewService(databaseFilename)
	if err != nil {
		return fmt.Errorf("failed to initialize subscription service: %w", err)
	}

	telegramBot, err := telegram.NewBot(m, s)
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	startSubscriptionCheck(m, s, telegramBot.Bot)

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
