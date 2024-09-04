package kimsufi

const (
	PlanCodeExample = "24ska01"
)

//
// Plan structure
//

type Plan struct {
	PlanCode      string            `json:"planCode"`
	InvoiceName   string            `json:"invoiceName"`
	AddonFamilies []PlanAddonFamily `json:"addonFamilies"`
	Product       string            `json:"product"`
	PricingType   string            `json:"pricingType"`
	//ConsumptionConfiguration string          `json:"consumptionConfiguration"`
	Pricings      []Pricing           `json:"pricings"`
	Configuration []PlanConfiguration `json:"configuration"`
	Family        string              `json:"family"`
	Blobs         PlanBlobs           `json:"blobs,omitempty"`
}

type PlanAddonFamily struct {
	Name      string   `json:"name"`
	Exclusive bool     `json:"exclusive"`
	Mandatory bool     `json:"mandatory"`
	Addons    []string `json:"addons"`
	Default   string   `json:"default"`
}

type Pricing struct {
	Phase           int               `json:"phase"`
	Capacities      []string          `json:"capacities"`
	Commitement     int               `json:"commitement"`
	Description     string            `json:"description"`
	Interval        int               `json:"interval"`
	IntervalUnit    string            `json:"intervalUnit"`
	Quantity        PlanPricingMinMax `json:"quantity"`
	Repeat          PlanPricingMinMax `json:"repeat"`
	Price           int               `json:"price"`
	Tax             int               `json:"tax"`
	Mode            string            `json:"mode"`
	Strategy        string            `json:"strategy"`
	MustBeCompleted bool              `json:"mustBeCompleted"`
	Type            string            `json:"type"`
	//Promotions []string `json:"promotions"`
	EngagementConfiguration PlanPricingEngagementConfiguration `json:"engagementConfiguration,omitempty"`
}

type PlanPricingMinMax struct {
	Max int `json:"max"`
	Min int `json:"min"`
}

type PlanPricingEngagementConfiguration struct {
	DefaultEndAction string `json:"defaultEndAction"`
	Duration         string `json:"duration"`
	Type             string `json:"type"`
}

type PlanConfiguration struct {
	Name        string   `json:"name"`
	IsCustom    bool     `json:"isCustom"`
	IsMandatory bool     `json:"isMandatory"`
	Values      []string `json:"values"`
}

type PlanBlobs struct {
	Commercial PlanBlobsCommercial `json:"commercial,omitempty"`
}

type PlanBlobsCommercial struct {
	Range string `json:"range,omitempty"`
}
