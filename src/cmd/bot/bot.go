package bot

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v3"
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
		log.Fatalf("failed to parse log-level: %v\n", err)
	}
	log.SetLevel(level)

	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return err
	}

	b.Handle("/hello", func(c tele.Context) error {
		username := c.Sender().Username

		return c.Send("Hello! @" + username)
	})

	b.Handle("/subscribe", func(c tele.Context) error {
		username := c.Sender().Username
		fmt.Printf("payload: %s\n", c.Message().Payload)
		var planCode string
		var datacentersString string
		n, err := fmt.Sscan(c.Message().Payload, &planCode, &datacentersString)
		if err != nil {
			fmt.Println("Invalid command")
			return err
		}
		var datacenters = strings.Split(strings.Trim(datacentersString, ","), ",")
		fmt.Printf("scanned %d variables: %s, %v\n", n, planCode, datacenters)

		var datacentersMessage string
		if len(datacenters) > 0 {
			datacentersMessage = "one of the following datacenters"
		} else {
			datacentersMessage = "this datacenter"
		}
		subscriptionId := 1

		return c.Send(fmt.Sprintf("@%s you will be notified when plan %s is available in %s %s (subscriptionId: %d)", username, planCode, datacentersMessage, strings.Join(datacenters, ", "), subscriptionId))
	})

	b.Handle("/unsubscribe", func(c tele.Context) error {
		username := c.Sender().Username
		fmt.Printf("payload: %s\n", c.Message().Payload)
		var subscriptionId int
		n, err := fmt.Sscan(c.Message().Payload, &subscriptionId)
		if err != nil {
			fmt.Println("Invalid command")
			return err
		}

		fmt.Printf("scanned %d variables: %d\n", n, subscriptionId)

		return c.Send(fmt.Sprintf("@%s you unsubscribed from subscription %d", username, subscriptionId))
	})

	b.Handle("/datacenters", func(c tele.Context) error {
		return c.Send("Available datacenters: ams, fra, nyc, sfo, tor")
	})

	b.Handle("/plans", func(c tele.Context) error {
		return c.Send("Available plans: 1cpu-1gb, 1cpu-2gb, 2cpu-4gb, 2cpu-8gb, 4cpu-8gb, 4cpu-16gb, 8cpu-16gb, 8cpu-32gb, 16cpu-32gb, 16cpu-64gb, 32cpu-64gb, 32cpu-128gb")
	})

	fmt.Println("Bot is running")
	b.Start()

	return nil
}
