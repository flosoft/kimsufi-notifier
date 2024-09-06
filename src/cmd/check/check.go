package check

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/kimsufi-notifier/cmd/flag"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
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
)

func init() {
	Cmd.PersistentFlags().StringSliceVarP(&datacenters, "datacenters", "d", nil, fmt.Sprintf("comma separated list of datacenters to check (allowed values: %s)", strings.Join(kimsufi.AllowedDatacenters, ", ")))
	Cmd.PersistentFlags().StringVarP(&planCode, "plan-code", "p", "", fmt.Sprintf("plan code name (e.g. %s)", kimsufi.PlanCodeExample))

}

func runner(cmd *cobra.Command, args []string) error {
	d := kimsufi.Config{
		URL:    kimsufi.GetOVHEndpoint(cmd.Flag(flag.OVHAPIEndpointFlagName).Value.String()),
		Logger: log.StandardLogger(),
	}
	m, err := kimsufi.NewService(d)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	k := m.Endpoint(cmd.Flag(flag.OVHAPIEndpointFlagName).Value.String())

	if planCode == "" {
		return fmt.Errorf("plan code is required")
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
			return fmt.Errorf("error: %w", err)
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
