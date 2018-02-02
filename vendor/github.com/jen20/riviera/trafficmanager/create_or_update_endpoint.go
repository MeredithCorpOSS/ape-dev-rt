package trafficmanager

import "github.com/jen20/riviera/azure"

type CreateOrUpdateTrafficManagerResponse struct {
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

type CreateOrUpdateTrafficManager struct {
	Name              string `json:"-" riviera:"name"`
	ResourceGroupName string `json:"-"`
	ProfileName       string `json:"-"`
	URIType           string `json:"-"`
	Type              string `json:"-" riviera:"type"`
	TargetResourceId  string `json:"targetResourceId,omitempty"`
	EndpointStatus    string `json:"endpointStatus,omitempty"`
	Weight            int    `json:"weight,omitempty"`
	Priority          int    `json:"priority,omitempty"`
	Target            string `json:"target,omitempty"`
	EndpointLocation  string `json:"endpointLocation,omitempty"`
	MinChildEndpoints int    `json:"minChildEndpoints,omitempty"`
}

func (s CreateOrUpdateTrafficManager) APIInfo() azure.APIInfo {
	return azure.APIInfo{
		APIVersion:  apiVersion,
		Method:      "PUT",
		URLPathFunc: trafficManagerEndpointURLPath(s.ResourceGroupName, s.ProfileName, s.URIType, s.Name),
		ResponseTypeFunc: func() interface{} {
			return &CreateOrUpdateTrafficManagerResponse{}
		},
	}
}
