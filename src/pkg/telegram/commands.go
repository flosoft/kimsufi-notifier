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
		"categories": {
			"/categories",
			"List available categories",
		},
		"countries": {
			"/countries",
			"List available countries",
		},
		"list": {
			"/list",
			"List available plans / servers",
		},
		"check": {
			"/check",
			"Check availability of a plan",
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
	}
)
