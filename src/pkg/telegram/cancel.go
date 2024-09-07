package telegram

import (
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) cancelWrapper(c tele.Context) error {
	log.Infof("cancelWrapper user=%s", formatUser(c.Sender()))

	c.Delete()
	return c.Respond(&tele.CallbackResponse{})
}
