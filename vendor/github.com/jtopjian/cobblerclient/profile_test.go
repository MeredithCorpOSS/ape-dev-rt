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

func TestGetProfiles(t *testing.T) {
	c := createStubHTTPClient(t, "get-profiles-req.xml", "get-profiles-res.xml")
	profiles, err := c.GetProfiles()
	utils.FailOnError(t, err)

	if len(profiles) != 1 {
		t.Errorf("Wrong number of profiles returned.")
	}
}

func TestGetProfile(t *testing.T) {
	c := createStubHTTPClient(t, "get-profile-req.xml", "get-profile-res.xml")
	profile, err := c.GetProfile("Ubuntu-14.04-x86_64")
	utils.FailOnError(t, err)

	if profile.Name != "Ubuntu-14.04-x86_64" {
		t.Errorf("Wrong profile returned.")
	}
}
