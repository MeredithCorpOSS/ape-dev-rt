/*
Copyright 2015 Container Solutions

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cobblerclient

import (
	"testing"

	"github.com/ContainerSolutions/go-utils"
)

func TestCreateKickstartFile(t *testing.T) {
	c := createStubHTTPClient(t, "create-kickstart-file-req.xml", "create-kickstart-file-res.xml")

	ks := KickstartFile{
		Name: "/var/lib/cobbler/kickstarts/foo.ks",
		Body: "sample content",
	}

	err := c.CreateKickstartFile(ks)
	utils.FailOnError(t, err)
}

func TestGetKickstartFile(t *testing.T) {
	ksName := "/var/lib/cobbler/kickstarts/foo.ks"

	c := createStubHTTPClient(t, "get-kickstart-file-req.xml", "get-kickstart-file-res.xml")

	expectedKS := KickstartFile{
		Name: ksName,
		Body: "sample content",
	}

	returnedKS, err := c.GetKickstartFile(ksName)
	utils.FailOnError(t, err)

	if returnedKS.Body != expectedKS.Body {
		t.Errorf("Kickstart Body did not match.")
	}
}
