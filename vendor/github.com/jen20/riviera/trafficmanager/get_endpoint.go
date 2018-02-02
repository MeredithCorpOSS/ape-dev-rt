package trafficmanager

import "github.com/jen20/riviera/azure"

type GetTrafficManagerEndpointResponse struct {
	ID                    *string `mapstructure:"id"`
	Name                  *string `mapstructure:"name"`
	Type                  *string `mapstructure:"type"`
	TargetResourceId      *string `mapstructure:"targetResourceId"`
	EndpointStatus        *string `mapstructure:"endpointStatus"`
	EndpointMonitorStatus *string `mapstructure:"endpointMonitorStatus"`
	Weight                int     `mapstructure:"weight"`
	Priority              int     `mapstructure:"priority"`
	Target                *string `mapstructure:"target"`
	EndpointLocation      *string `mapstructure:"endpointLocation"`
	MinChildEndpoints     int     `mapstructure:"minChildEndpoints"`
}

type GetTrafficManagerEndpoint struct {
	Name              string `json:"-"`
	URIType           string `json:"-"`
	ProfileName       string `json:"-"`
	ResourceGroupName string `json:"-"`
}

func (s GetTrafficManagerEndpoint) APIInfo() azure.APIInfo {
	return azure.APIInfo{
		APIVersion:  apiVersion,
		Method:      "GET",
		URLPathFunc: trafficManagerEndpointURLPath(s.ResourceGroupName, s.ProfileName, s.URIType, s.Name),
		ResponseTypeFunc: func() interface{} {
			return &CreateOrUpdateTrafficManagerResponse{}
		},
	}
}
