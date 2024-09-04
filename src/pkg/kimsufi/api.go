package kimsufi

import "github.com/ovh/go-ovh/ovh"

const (
	DefaultOVHAPIEndpointName = "ovh-eu"
)

func AllOVHAPIEndpointsNames() (endpointsNames []string) {
	for name, _ := range ovh.Endpoints {
		endpointsNames = append(endpointsNames, name)
	}
	return
}

func GetOVHEndpoint(name string) string {
	if _, ok := ovh.Endpoints[name]; ok {
		return ovh.Endpoints[name]
	}
	return ""
}
