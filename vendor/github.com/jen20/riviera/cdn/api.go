package cdn

import "fmt"

const apiVersion = "2015-06-01"
const apiProvider = "Microsoft.Cdn"

func cdnProfileDefaultURLPath(resourceGroupName, profileName string) func() string {
	return func() string {
		return fmt.Sprintf("resourceGroups/%s/providers/%s/Profiles/%s", resourceGroupName, apiProvider, profileName)
	}
}
