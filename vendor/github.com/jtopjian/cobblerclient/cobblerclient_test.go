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
	"regexp"
	"testing"

	"github.com/ContainerSolutions/go-utils"
)

var config = ClientConfig{
	Url:      "http://example.org:1234",
	Username: "john",
	Password: "doe",
}

func createStubHTTPClient(t *testing.T, reqFixture string, resFixture string) Client {
	hc := utils.NewStubHTTPClient(t)

	if reqFixture != "" {
		rawRequest, err := utils.Fixture(reqFixture)
		utils.FailOnError(t, err)

		// flatten the request so it matches the kolo generated xml
		r := regexp.MustCompile(`\s+<`)
		expectedReq := []byte(r.ReplaceAllString(string(rawRequest), "<"))
		hc.Expected = expectedReq
	}

	if resFixture != "" {
		response, err := utils.Fixture(resFixture)
		utils.FailOnError(t, err)
		hc.Response = response
	}

	c := NewClient(hc, config)
	c.Token = "securetoken99"
	return c
}

func TestLogin(t *testing.T) {
	c := createStubHTTPClient(t, "login-req.xml", "login-res.xml")
	ok, err := c.Login()
	utils.FailOnError(t, err)

	if !ok {
		t.Errorf("true expected; got false")
	}

	expected := "sa/1EWr40BWU+Pq3VEOOpD4cQtxkeMuFUw=="
	if c.Token != expected {
		t.Errorf(`"%s" expected; got "%s"`, expected, c.Token)
	}
}

func TestLoginWithError(t *testing.T) {
	c := createStubHTTPClient(t, "login-req.xml", "login-res-err.xml")
	expected := `error: "<class 'cobbler.cexceptions.CX'>:'login failed (cobbler)'" code: 1`

	ok, err := c.Login()
	if ok {
		t.Errorf("false expected; got true")
	}

	if err.Error() != expected {
		t.Errorf("%s expected; got %s", expected, err)
	}
}

func TestSync(t *testing.T) {
	c := createStubHTTPClient(t, "sync-req.xml", "sync-res.xml")
	expected := true

	result, err := c.Sync()
	utils.FailOnError(t, err)

	if result != expected {
		t.Errorf("%s expected; got %s", expected, result)
	}
}
