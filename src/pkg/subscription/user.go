package subscription

import tele "gopkg.in/telebot.v3"

type User struct {
	UserID int64
	User   *tele.User
}
