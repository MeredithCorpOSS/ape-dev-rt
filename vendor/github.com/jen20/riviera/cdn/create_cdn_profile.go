package cdn

import "github.com/jen20/riviera/azure"

type Sku struct {
	Name string `json:"name" mapstructure:"name"`
}

type CreateCDNProfileResponse struct {
	ID                string             `mapstructure:"id"`
	Name              string             `mapstructure:"name"`
	Location          string             `mapstructure:"location"`
	Tags              map[string]*string `mapstructure:"tags"`
	Sku               Sku                `mapstructure:"sku"`
	ProvisioningState string             `mapstructure:"provisioningState"`
	ResourceState     string             `mapstructure:"resourceState"`
}

type CreateCDNProfile struct {
	Name              string             `json:"-"`
	ResourceGroupName string             `json:"-"`
	Location          string             `json:"-" riviera:"location"`
	Tags              map[string]*string `json:"-" riviera:"tags"`
	Sku               Sku                `json:"-" riviera:"sku"`
}

func (command CreateCDNProfile) APIInfo() azure.APIInfo {
	return azure.APIInfo{
		APIVersion:  apiVersion,
		Method:      "PUT",
		URLPathFunc: cdnProfileDefaultURLPath(command.ResourceGroupName, command.Name),
		ResponseTypeFunc: func() interface{} {
			return &CreateCDNProfileResponse{}
		},
	}
}
