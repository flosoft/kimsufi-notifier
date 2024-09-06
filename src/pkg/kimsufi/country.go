package kimsufi

var (
	AllowedRegions = []Region{
		{
			Name:      "Europe",
			Endpoint:  "ovh-eu",
			Countries: []string{"CZ", "DE", "ES", "FI", "FR", "GB", "IE", "IT", "LT", "MA", "NL", "PL", "PT", "SN", "TN"},
		},
		{
			Name:      "Other",
			Endpoint:  "ovh-ca",
			Countries: []string{"ASIA", "AU", "CA", "IN", "QC", "SG", "WE", "WS"},
		},
		{
			Name:      "US",
			Endpoint:  "ovh-us",
			Countries: []string{"US"},
		},
	}
)

type Region struct {
	Name      string
	Endpoint  string
	Countries []string
}

type Order struct {
	Components Components `json:"components"`
}

type Components struct {
	Schemas Schema `json:"schemas"`
}

type Schema struct {
	DedicatedServerIpCountryEnum Component `json:"dedicated.server.IpCountryEnum"`
	OVHSubsidiaryEnum            Component `json:"nichandle.OvhSubsidiaryEnum"`
}

type Component struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	ENUM        []string `json:"enum"`
}

func (o *Order) GetCountries() []string {
	return o.Components.Schemas.OVHSubsidiaryEnum.ENUM
}
