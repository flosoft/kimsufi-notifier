package kimsufi

type Catalog struct {
	CatalogID    int          `json:"catalogId"`
	Locale       Locale       `json:"locale"`
	Plans        []Plan       `json:"plans"`
	Products     []Product    `json:"products"`
	Addons       []Addon      `json:"addons"`
	PlanFamilies []PlanFamily `json:"planFamilies"`
}

type Locale struct {
	CurrencyCode string `json:"currencyCode"`
	Subsidiary   string `json:"subsidiary"`
	TaxRate      int    `json:"taxRate"`
}

//
// Product structure
//

type Product struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Blobs       ProductBlobs `json:"blobs"`
}

type ProductBlobs struct {
	Technical ProductBlobsTechnical `json:"technical"`
}

type ProductBlobsTechnical struct {
	Storage ProductBlobsTechnicalStorage `json:"storage"`
}

type ProductBlobsTechnicalStorage struct {
	Raid        string                                  `json:"raid"`
	Disks       []ProductBlobsTechnicalStorageDisk      `json:"disks"`
	HotSwap     bool                                    `json:"hotSwap"`
	RaidDetails ProductBlobsTechnicalStorageRaidDetails `json:"raidDetails"`
}

type ProductBlobsTechnicalStorageDisk struct {
	Specs      string `json:"specs"`
	Usage      string `json:"usage"`
	Number     int    `json:"number"`
	Capacity   int    `json:"capacity"`
	Interface  string `json:"interface"`
	Technology string `json:"technology"`
}

type ProductBlobsTechnicalStorageRaidDetails struct {
	Type string `json:"type"`
}

//
// Addon structure
//

type Addon struct {
	PlanCode    string `json:"planCode"`
	InvoiceName string `json:"invoiceName"`
	// AddonFamilies []AddonAddonFamily `json:"addonFamilies"`
	Product     string `json:"product"`
	PricingType string `json:"pricingType"`
	// ConsumptionConfiguration string          `json:"consumptionConfiguration"`
	Pricings []Pricing `json:"pricings"`
	//Configurations []AddonConfiguration `json:"configurations"`
	//Family string `json:"family"`
	//Blobs AddonBlobs `json:"blobs"`
}

//
// PlanFamily structure
//

type PlanFamily struct {
	Name string `json:"name"`
}
