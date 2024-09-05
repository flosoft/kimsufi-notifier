package telegram

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
