package list

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/kimsufi-notifier/cmd/flag"
	"github.com/TheoBrigitte/kimsufi-notifier/pkg/kimsufi"
)

var (
	Cmd = &cobra.Command{
		Use:   "list",
		Short: "list available servers",
		RunE:  runner,
	}

	datacenters   []string
	planCode      string
	ovhSubsidiary string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&ovhSubsidiary, "country", "c", "FR", "country code to filter entries")
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
		return fmt.Errorf("failed to initialize kimsufi service: %w", err)
	}
	k := m.Endpoint(cmd.Flag(flag.OVHAPIEndpointFlagName).Value.String())

	c, err := k.ListServers(ovhSubsidiary)
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	a, err := k.GetAvailabilities(datacenters, planCode)
	if err != nil {
		return fmt.Errorf("failed to list availabilities: %w", err)
	}

	log.Infof("Found %d plans\n", len(c.Plans))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "planCode\tcategory\tname\tprice\tstatus\tdatacenters")
	fmt.Fprintln(w, "--------\t--------\t----\t-----\t------\t-----------")

	sort.Slice(c.Plans, func(i, j int) bool {
		return c.Plans[i].FirstPrice().Price < c.Plans[j].FirstPrice().Price
	})

	var nothingAvailable bool = true
	for _, category := range kimsufi.PlanCategories {
		for _, plan := range c.Plans {
			if planCode != "" && plan.PlanCode != planCode {
				continue
			}
			if plan.Blobs.Commercial.Range != category {
				continue
			}

			var price float64
			planPrice := plan.FirstPrice()
			if !reflect.DeepEqual(planPrice, kimsufi.Pricing{}) {
				price = float64(planPrice.Price) / kimsufi.PriceDivider
			}

			var status string
			datacenters := a.GetPlanCodeAvailableDatacenters(plan.PlanCode)
			if len(datacenters) == 0 {
				status = kimsufi.StatusUnavailable
			} else {
				nothingAvailable = false
				status = kimsufi.StatusAvailable
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%s\t%s\n", plan.PlanCode, category, plan.InvoiceName, price, status, strings.Join(datacenters, ", "))
		}
	}
	w.Flush()

	if nothingAvailable {
		log.Warnf("no server available\n")
		os.Exit(1)
	}

	return nil
}
