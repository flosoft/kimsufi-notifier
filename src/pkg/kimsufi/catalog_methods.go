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

func (c Catalog) PlanExists(planCode string) bool {
	for _, plan := range c.Plans {
		if plan.PlanCode == planCode {
			return true
		}
	}

	return false
}
