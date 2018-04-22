package deploymentstate

import (
	"fmt"
	"testing"

	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/backends/backendstest"
	"github.com/TimeIncOSS/ape-dev-rt/hcl"
)

func TestLoadBackends(t *testing.T) {
	supportedBackends["fixture"] = &backendstest.FixtureBackend{}

	var emptyVars interface{}
	cases := []struct {
		Path          string
		Vars          interface{}
		ExpectedError error
	}{
		0: {"test-fixtures/no-deployment-state.hcl", emptyVars,
			fmt.Errorf(`Failed to load config from "test-fixtures/no-deployment-state.hcl": Unrecognised config block ("random_thing_oink"), supported: ["deployment_state" "remote_state"]`)},
		1: {"test-fixtures/unexpected-resource.hcl", emptyVars,
			fmt.Errorf(`Failed to load config from "test-fixtures/unexpected-resource.hcl": Unrecognised config block ("random_thing_oink"), supported: ["deployment_state" "remote_state"]`)},
		2: {"test-fixtures/empty-file.hcl", emptyVars,
			fmt.Errorf("No configuration provided")},
		3: {"test-fixtures/uninitializable-backend.hcl", emptyVars,
			fmt.Errorf(`Error initializing backend: "Unable to initalize backend with config: map[\"key\":\"//yada/\"]: Missing bucket field in config"`)},
		4: {"test-fixtures/unsupported-backend.hcl", emptyVars,
			fmt.Errorf("Defined backend something-unsupported is not supported")},
		5: {"test-fixtures/double-resource.hcl", emptyVars,
			fmt.Errorf("Duplicate backend defined (fixture)")},
		6: {"test-fixtures/valid.hcl", emptyVars, nil},
	}

	for i, c := range cases {
		err := testLoadBackendsFixture(c.Path)
		if err == nil || c.ExpectedError == nil {
			if err != c.ExpectedError {
				t.Fatalf("%d: Errors don't match.\nExpected: %s\n   Given: %s\n", i, c.ExpectedError, err)
			}
			continue
		}

		if err.Error() != c.ExpectedError.Error() {
			t.Fatalf("%d: Errors don't match.\nExpected: %s\n   Given: %s\n", i, c.ExpectedError, err)
		}
	}
}

func testLoadBackendsFixture(path string) error {
	cfg, _, err := hcl.LoadConfigFromPath("", "", path)
	if err != nil {
		return err
	}

	_, err = New(cfg.DeploymentState)
	return err
}
