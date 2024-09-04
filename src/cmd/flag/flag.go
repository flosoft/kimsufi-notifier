package flag

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/logger"
)

const (
	LogLevelFlagName       = "log-level"
	OVHAPIEndpointFlagName = "ovh-api-endpoint"
)

func Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(LogLevelFlagName, "l", log.ErrorLevel.String(), fmt.Sprintf("log level (allowed values: %s)", strings.Join(logger.AllLevelsString(), ", ")))
	cmd.PersistentFlags().String(OVHAPIEndpointFlagName, kimsufi.DefaultOVHAPIEndpointName, fmt.Sprintf("OVH API Endpoint (allowed values: %s)", strings.Join(kimsufi.AllOVHAPIEndpointsNames(), ", ")))
}
