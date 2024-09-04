package check

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ovh/go-ovh/ovh"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/logger"
)

// Cmd represents the check command
var (
	Cmd = &cobra.Command{
		Use:   "check",
		Short: "check availability",
		RunE:  runner,
	}

	datacenters []string
	planCode    string

	logLevel string
)

const (
	kimsufiAPI = ovh.OvhEU
	smsAPI     = "https://smsapi.free-mobile.fr/sendmsg"
)

func init() {
	Cmd.PersistentFlags().StringSliceVarP(&datacenters, "datacenters", "d", nil, fmt.Sprintf("comma separated list of datacenters to check (allowed values: %s)", strings.Join(kimsufi.AllowedDatacenters, ", ")))
	Cmd.PersistentFlags().StringVarP(&planCode, "plan-code", "p", "", fmt.Sprintf("plan code name (e.g. %s)", kimsufi.PlanCodeExample))
	Cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "error", fmt.Sprintf("log level (allowed values: %s)", strings.Join(logger.AllLevelsString(), ", ")))

}

func runner(cmd *cobra.Command, args []string) error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("failed to parse log-level: %v\n", err)
	}

	log.SetLevel(level)
	d := kimsufi.Config{
		URL:    kimsufiAPI,
		Logger: log.StandardLogger(),
	}
	k, err := kimsufi.NewService(d)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	if planCode == "" {
		log.Fatalf("plan code is required\n")
	}

	a, err := k.GetAvailabilities(datacenters, planCode)
	if err != nil {
		if kimsufi.IsNotAvailableError(err) {
			datacenterMessage := ""
			if len(datacenters) > 0 {
				datacenterMessage = strings.Join(datacenters, ", ")
			} else {
				datacenterMessage = "all datacenters"
			}
			log.Printf("%s is not available in %s\n", planCode, datacenterMessage)
			return nil
		} else {
			log.Fatalf("error: %v\n", err)
		}
	}

	formatter := kimsufi.DatacenterFormatter(kimsufi.IsDatacenterAvailable, kimsufi.DatacenterKey)
	result := a.Format(kimsufi.PlanCode, formatter)
	//data, err := json.Marshal(result)
	//log.Printf("%s\n", data)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "planCode\tstatus\tdatacenters")
	fmt.Fprintln(w, "--------\t------\t-----------")

	for k, v := range result {
		status := "available"
		if len(v) == 0 {
			status = "unavailable"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", k, status, strings.Join(v, ", "))
	}

	w.Flush()
	//for _, availability := range *a {
	//	if availability.IsAvailable() {
	//		planDatacenters := formatter(availability.Datacenters)
	//		fmt.Printf("%s is available in following datacenters %v\n", availability.PlanCode, planDatacenters)
	//	} else {
	//		fmt.Printf("%s is not available\n", availability.PlanCode)
	//	}
	//}

	return nil
}
