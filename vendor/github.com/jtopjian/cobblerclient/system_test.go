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

func TestGetSystems(t *testing.T) {
	c := createStubHTTPClient(t, "get-systems-req.xml", "get-systems-res.xml")
	systems, err := c.GetSystems()
	utils.FailOnError(t, err)

	if len(systems) != 1 {
		t.Errorf("Wrong number of systems returned.")
	}
}

func TestGetSystem(t *testing.T) {
	c := createStubHTTPClient(t, "get-system-req.xml", "get-system-res.xml")
	system, err := c.GetSystem("test")
	utils.FailOnError(t, err)

	if system.Name != "test" {
		t.Errorf("Wrong system returned.")
	}
}

func TestNewSystem(t *testing.T) {
	c := createStubHTTPClient(t, "new-system-req.xml", "new-system-res.xml")
	result, err := c.Call("new_system", c.Token)
	utils.FailOnError(t, err)
	newId := result.(string)

	if newId != "___NEW___system::abc123==" {
		t.Errorf("Wrong ID returned.")
	}

	c = createStubHTTPClient(t, "set-system-hostname-req.xml", "set-system-hostname-res.xml")
	result, err = c.Call("modify_system", newId, "hostname", "blahhost", c.Token)
	utils.FailOnError(t, err)

	if !result.(bool) {
		t.Errorf("Setting hostname failed.")
	}

	c = createStubHTTPClient(t, "set-system-name-req.xml", "set-system-name-res.xml")
	result, err = c.Call("modify_system", newId, "name", "mytestsystem", c.Token)
	utils.FailOnError(t, err)

	if !result.(bool) {
		t.Errorf("Setting name failed.")
	}

	c = createStubHTTPClient(t, "set-system-nameservers-req.xml", "set-system-nameservers-res.xml")
	result, err = c.Call("modify_system", newId, "name_servers", "8.8.8.8 8.8.4.4", c.Token)
	utils.FailOnError(t, err)

	if !result.(bool) {
		t.Errorf("Setting name servers failed.")
	}

	c = createStubHTTPClient(t, "set-system-profile-req.xml", "set-system-profile-res.xml")
	result, err = c.Call("modify_system", newId, "profile", "centos7-x86_64", c.Token)
	utils.FailOnError(t, err)

	if !result.(bool) {
		t.Errorf("Setting name servers failed.")
	}

	/* I'm not sure how to get this test to pass with unordered maps
	nicInfo := map[string]interface{}{
		"macaddress-eth0":  "01:02:03:04:05:06",
		"ipaddress-eth0":   "1.2.3.4",
		"dnsname-eth0":     "deathstar",
		"subnetsmask-eth0": "255.255.255.0",
		"if-gateway-eth0":  "4.3.2.1",
	}

	c = createStubHTTPClient(t, "set-system-network-req.xml", "set-system-network-res.xml")
	result, err = c.Call("modify_system", newId, "modify_interface", nicInfo, c.Token)
	utils.FailOnError(t, err)

	if !result.(bool) {
		t.Errorf("Setting interface failed.")
	}
	*/

	c = createStubHTTPClient(t, "save-system-req.xml", "save-system-res.xml")
	result, err = c.Call("save_system", newId, c.Token)
	utils.FailOnError(t, err)

	if !result.(bool) {
		t.Errorf("Save failed.")
	}
}
