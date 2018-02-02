package trafficmanager

import "fmt"

const apiVersion = "2015-11-01"
const apiProvider = "Microsoft.Network"

func trafficManagerProfileURLPath(resourceGroupName, profileName string) func() string {
	return func() string {
		return fmt.Sprintf("resourceGroups/%s/providers/%s/trafficManagerProfiles/%s", resourceGroupName, apiProvider, profileName)
	}
}

func trafficManagerEndpointURLPath(resourceGroupName, profileName, endpointType, endpointName string) func() string {
	return func() string {
		return fmt.Sprintf("resourceGroups/%s/providers/%s/trafficManagerProfiles/%s/%s/%s", resourceGroupName, apiProvider, profileName, endpointType, endpointName)
	}
}
