package cdn

import "github.com/jen20/riviera/azure"

type GetCDNProfileResponse struct {
	ID                *string            `mapstructure:"id"`
	Name              string             `mapstructure:"name"`
	ResourceGroupName string             `mapstructure:"-"`
	Location          string             `mapstructure:"location"`
	Tags              map[string]*string `mapstructure:"tags"`
	Sku               Sku                `mapstructure:"sku"`
	ProvisioningState string             `mapstructure:"provisioningState"`
	ResourceState     string             `mapstructure:"resourceState"`
}

type GetCDNProfile struct {
	Name              string `json:"-"`
	ResourceGroupName string `json:"-"`
}

func (s GetCDNProfile) APIInfo() azure.APIInfo {
	return azure.APIInfo{
		APIVersion:  apiVersion,
		Method:      "GET",
		URLPathFunc: cdnProfileDefaultURLPath(s.ResourceGroupName, s.Name),
		ResponseTypeFunc: func() interface{} {
			return &GetCDNProfileResponse{}
		},
	}
}
