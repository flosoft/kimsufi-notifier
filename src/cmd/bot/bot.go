package bot

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/kimsufi-notifier/cmd/flag"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
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
		"help":       "/help",
		"categories": "/categories",
		"countries":  "/countries",
		"list":       "/list",
		"check":      "/check",
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

	telegramBot, err := telegram.NewBot()
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	telegramBot.Handle("/help", helpCommand)
	telegramBot.Handle("/categories", categoriesCommand)
	telegramBot.Handle("/countries", countriesCommand)
	telegramBot.Handle("/list", listCommand(k))
	telegramBot.Handle("/check", checkCommand(k))

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
