package cdn

import "github.com/jen20/riviera/azure"

type UpdateCDNProfileTagsResponse struct {
	ID                string             `mapstructure:"id"`
	Name              string             `mapstructure:"name"`
	Location          string             `mapstructure:"location"`
	Tags              map[string]*string `mapstructure:"tags"`
	Sku               Sku                `mapstructure:"sku"`
	ProvisioningState string             `mapstructure:"provisioningState"`
	ResourceState     string             `mapstructure:"resourceState"`
}

type UpdateCDNProfileTags struct {
	Name              string             `json:"-"`
	ResourceGroupName string             `json:"-"`
	Tags              map[string]*string `json:"-" riviera:"tags"`
}

func (command UpdateCDNProfileTags) APIInfo() azure.APIInfo {
	return azure.APIInfo{
		APIVersion:  apiVersion,
		Method:      "PATCH",
		URLPathFunc: cdnProfileDefaultURLPath(command.ResourceGroupName, command.Name),
		ResponseTypeFunc: func() interface{} {
			return &UpdateCDNProfileTagsResponse{}
		},
	}
}
