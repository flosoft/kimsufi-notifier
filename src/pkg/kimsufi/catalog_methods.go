package kimsufi

import "math"

var (
	PriceDecimals = 8
	PriceDivider  = math.Pow10(PriceDecimals)

	// List of known plan categories, including an empty string for uncategorized plans
	PlanCategories = []string{"kimsufi", "soyoustart", "rise", ""}

	StatusAvailable   = "available"
	StatusUnavailable = "unavailable"
)

func (p Plan) FirstPrice() Pricing {
	if len(p.Pricings) == 0 {
		return Pricing{}
	}

	for _, price := range p.Pricings {
		if price.Phase == 1 && price.Mode == "default" {
			return price
		}
	}

	return p.Pricings[0]
}

func (c Catalog) GetPlan(planCode string) *Plan {
	for _, plan := range c.Plans {
		if plan.PlanCode == planCode {
			return &plan
		}
	}

	return nil
}

func (c Catalog) PlanExists(planCode string) bool {
	for _, plan := range c.Plans {
		if plan.PlanCode == planCode {
			return true
		}
	}

	return false
}

func (p Plan) GetDatacenters() []string {
	for _, config := range p.Configurations {
		if config.Name == "dedicated_datacenter" {
			return config.Values
		}
	}

	return nil
}
