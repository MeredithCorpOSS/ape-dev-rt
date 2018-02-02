package cdn

import "github.com/jen20/riviera/azure"

type DeleteCDNProfile struct {
	Name              string `json:"-"`
	ResourceGroupName string `json:"-"`
}

func (s DeleteCDNProfile) APIInfo() azure.APIInfo {
	return azure.APIInfo{
		APIVersion:  apiVersion,
		Method:      "DELETE",
		URLPathFunc: cdnProfileDefaultURLPath(s.ResourceGroupName, s.Name),
		ResponseTypeFunc: func() interface{} {
			return nil
		},
	}
}
