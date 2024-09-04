package bot

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/telegram"
)

var (
	Cmd = &cobra.Command{
		Use:   "bot",
		Short: "Run Telegram bot",
		RunE:  runner,
	}

	datacenters   []string
	logLevel      string
	planCode      string
	ovhSubsidiary string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "error", "log level (allowed values: debug, info, warn, error, fatal, panic)")
}

func runner(cmd *cobra.Command, args []string) error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log-level: %w", err)
	}
	log.SetLevel(level)

	telegramBot, err := telegram.NewBot()
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
