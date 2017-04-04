package acl_test

import (
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestGetChecks(t *testing.T) {
	tests := []struct {
		name    string
		acl     acl.ACL
		priv    acl.Privileges
		granted bool
	}{
		{
			name:    "none",
			acl:     acl.ACL{},
			priv:    acl.NoPrivs,
			granted: false,
		},
		{
			name: "insert",
			acl: acl.ACL{
				GrantOptions: acl.Insert,
				Privileges:   acl.Insert,
			},
			priv:    acl.Insert,
			granted: true,
		},
		{
			name: "select",
			acl: acl.ACL{
				GrantOptions: acl.Select,
				Privileges:   acl.Select,
			},
			priv:    acl.Select,
			granted: true,
		},
		{
			name: "update",
			acl: acl.ACL{
				GrantOptions: acl.Update,
				Privileges:   acl.Update,
			},
			priv:    acl.Update,
			granted: true,
		},
		{
			name: "delete",
			acl: acl.ACL{
				GrantOptions: acl.Delete,
				Privileges:   acl.Delete,
			},
			priv:    acl.Delete,
			granted: true,
		},
		{
			name: "truncate",
			acl: acl.ACL{
				GrantOptions: acl.Truncate,
				Privileges:   acl.Truncate,
			},
			priv:    acl.Truncate,
			granted: true,
		},
		{
			name: "references",
			acl: acl.ACL{
				GrantOptions: acl.References,
				Privileges:   acl.References,
			},
			priv:    acl.References,
			granted: true,
		},
		{
			name: "trigger",
			acl: acl.ACL{
				GrantOptions: acl.Trigger,
				Privileges:   acl.Trigger,
			},
			priv:    acl.Trigger,
			granted: true,
		},
		{
			name: "execute",
			acl: acl.ACL{
				GrantOptions: acl.Execute,
				Privileges:   acl.Execute,
			},
			priv:    acl.Execute,
			granted: true,
		},
		{
			name: "usage",
			acl: acl.ACL{
				GrantOptions: acl.Usage,
				Privileges:   acl.Usage,
			},
			priv:    acl.Usage,
			granted: true,
		},
		{
			name: "create",
			acl: acl.ACL{
				GrantOptions: acl.Create,
				Privileges:   acl.Create,
			},
			priv:    acl.Create,
			granted: true,
		},
		{
			name: "temporary",
			acl: acl.ACL{
				GrantOptions: acl.Temporary,
				Privileges:   acl.Temporary,
			},
			priv:    acl.Temporary,
			granted: true,
		},
		{
			name: "connect",
			acl: acl.ACL{
				GrantOptions: acl.Connect,
				Privileges:   acl.Connect,
			},
			priv:    acl.Connect,
			granted: true,
		},
	}

	for i, test := range tests {
		if test.name == "" {
			t.Fatalf("test %d name empty", i)
		}

		t.Run(test.name, func(t *testing.T) {
			if test.acl.GetGrantOption(test.priv) != test.granted {
				t.Fatalf("expected %t", test.granted)
			}
			if test.acl.GetPrivilege(test.priv) != test.granted {
				t.Fatalf("expected %t", test.granted)
			}
		})
	}
}
