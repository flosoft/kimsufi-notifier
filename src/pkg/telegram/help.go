package telegram

import (
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func helpCommand(c tele.Context) error {
	log.Info("Handle /help command user=" + formatUser(c.Sender()))

	output := "Hello,\n"
	output += "This bot can help you to monitor the availability of Kimsufi servers.\n"
	output += "\n"
	output += "You can subscribe to a plan and get notified when it becomes available with /subscribe command.\n"
	output += "You can also list your current subscriptions /listsubscriptions command.\n"
	output += "\n"
	output += "You can use the following commands:\n"

	commandsIndex := []string{"help", "subscribe", "unsubscribe", "listsubscriptions", "hello"}
	for _, ci := range commandsIndex {
		output += commands[ci].command + "  " + commands[ci].help + "\n"
	}

	output += "\n"
	output += "Ask for support or report issues in our Telegram group: https://t.me/+xPnf7KSGEoA1Nzcy\n"

	return c.Send(output, tele.ModeHTML)
}
