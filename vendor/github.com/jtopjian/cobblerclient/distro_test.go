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

func TestGetDistros(t *testing.T) {
	c := createStubHTTPClient(t, "get-distros-req.xml", "get-distros-res.xml")
	distros, err := c.GetDistros()
	utils.FailOnError(t, err)

	if len(distros) != 1 {
		t.Errorf("Wrong number of distros returned.")
	}
}

func TestGetDistro(t *testing.T) {
	c := createStubHTTPClient(t, "get-distro-req.xml", "get-distro-res.xml")
	distro, err := c.GetDistro("Ubuntu-14.04-x86_64")
	utils.FailOnError(t, err)

	if distro.Name != "Ubuntu-14.04-x86_64" {
		t.Errorf("Wrong distro returned.")
	}
}

/*
 * NOTE: We're skipping the testing of CREATE, UPDATE, DELETE methods for now because
 *       the current implementation of the StubHTTPClient does not allow
 *       buffered mock responses so as soon as the method makes the second
 *       call to Cobbler it'll fail.
 *       This is a system test, so perhaps we can run Cobbler in a Docker container
 *       and take it from there.
 */
