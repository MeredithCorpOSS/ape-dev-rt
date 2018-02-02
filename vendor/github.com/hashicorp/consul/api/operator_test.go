package api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/consul/testutil"
)

func TestOperator_RaftGetConfiguration(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	operator := c.Operator()
	out, err := operator.RaftGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out.Servers) != 1 ||
		!out.Servers[0].Leader ||
		!out.Servers[0].Voter {
		t.Fatalf("bad: %v", out)
	}
}

func TestOperator_RaftRemovePeerByAddress(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	// If we get this error, it proves we sent the address all the way
	// through.
	operator := c.Operator()
	err := operator.RaftRemovePeerByAddress("nope", nil)
	if err == nil || !strings.Contains(err.Error(),
		"address \"nope\" was not found in the Raft configuration") {
		t.Fatalf("err: %v", err)
	}
}

func TestOperator_KeyringInstallListPutRemove(t *testing.T) {
	oldKey := "d8wu8CSUrqgtjVsvcBPmhQ=="
	newKey := "qxycTi/SsePj/TZzCBmNXw=="
	t.Parallel()
	c, s := makeClientWithConfig(t, nil, func(c *testutil.TestServerConfig) {
		c.Encrypt = oldKey
	})
	defer s.Stop()

	operator := c.Operator()
	if err := operator.KeyringInstall(newKey, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	listResponses, err := operator.KeyringList(nil)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	// Make sure the new key is installed
	if len(listResponses) != 2 {
		t.Fatalf("bad: %v", len(listResponses))
	}
	for _, response := range listResponses {
		if len(response.Keys) != 2 {
			t.Fatalf("bad: %v", len(response.Keys))
		}
		if _, ok := response.Keys[oldKey]; !ok {
			t.Fatalf("bad: %v", ok)
		}
		if _, ok := response.Keys[newKey]; !ok {
			t.Fatalf("bad: %v", ok)
		}
	}

	// Switch the primary to the new key
	if err := operator.KeyringUse(newKey, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := operator.KeyringRemove(oldKey, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	listResponses, err = operator.KeyringList(nil)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	// Make sure the old key is removed
	if len(listResponses) != 2 {
		t.Fatalf("bad: %v", len(listResponses))
	}
	for _, response := range listResponses {
		if len(response.Keys) != 1 {
			t.Fatalf("bad: %v", len(response.Keys))
		}
		if _, ok := response.Keys[oldKey]; ok {
			t.Fatalf("bad: %v", ok)
		}
		if _, ok := response.Keys[newKey]; !ok {
			t.Fatalf("bad: %v", ok)
		}
	}
}

func TestOperator_AutopilotGetSetConfiguration(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	operator := c.Operator()
	config, err := operator.AutopilotGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %v", config)
	}

	// Change a config setting
	newConf := &AutopilotConfiguration{CleanupDeadServers: false}
	if err := operator.AutopilotSetConfiguration(newConf, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	config, err = operator.AutopilotGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if config.CleanupDeadServers {
		t.Fatalf("bad: %v", config)
	}
}

func TestOperator_AutopilotCASConfiguration(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	operator := c.Operator()
	config, err := operator.AutopilotGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %v", config)
	}

	// Pass an invalid ModifyIndex
	{
		newConf := &AutopilotConfiguration{
			CleanupDeadServers: false,
			ModifyIndex:        config.ModifyIndex - 1,
		}
		resp, err := operator.AutopilotCASConfiguration(newConf, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if resp {
			t.Fatalf("bad: %v", resp)
		}
	}

	// Pass a valid ModifyIndex
	{
		newConf := &AutopilotConfiguration{
			CleanupDeadServers: false,
			ModifyIndex:        config.ModifyIndex,
		}
		resp, err := operator.AutopilotCASConfiguration(newConf, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !resp {
			t.Fatalf("bad: %v", resp)
		}
	}
}

func TestOperator_ServerHealth(t *testing.T) {
	t.Parallel()
	c, s := makeClientWithConfig(t, nil, func(c *testutil.TestServerConfig) {
		c.RaftProtocol = 3
	})
	defer s.Stop()

	operator := c.Operator()
	testutil.WaitForResult(func() (bool, error) {
		out, err := operator.AutopilotServerHealth(nil)
		if err != nil {
			return false, fmt.Errorf("err: %v", err)
		}
		if len(out.Servers) != 1 ||
			!out.Servers[0].Healthy ||
			out.Servers[0].Name != s.Config.NodeName {
			return false, fmt.Errorf("bad: %v", out)
		}

		return true, nil
	}, func(err error) {
		t.Fatal(err)
	})
}
