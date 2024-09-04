package bot

import (
	"fmt"

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
	planCode      string
	ovhSubsidiary string
)

func init() {
}

func runner(cmd *cobra.Command, args []string) error {
	telegramBot, err := telegram.NewBot()
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	fmt.Println("Bot is running")
	telegramBot.Start()

	return nil
}
