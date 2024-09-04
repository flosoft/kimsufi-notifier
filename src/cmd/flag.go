package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/logger"
)

func init() {
	rootCmd.PersistentFlags().StringP("log-level", "l", log.ErrorLevel.String(), fmt.Sprintf("log level (allowed values: %s)", strings.Join(logger.AllLevelsString(), ", ")))
}
