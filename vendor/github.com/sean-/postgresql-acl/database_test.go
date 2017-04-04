package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestDatabaseString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		want acl.Database
		fail bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.Database{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
		},
		{
			name: "all without grant",
			in:   "foo=CTc",
			out:  "foo=CTc",
			want: acl.Database{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Create | acl.Temporary | acl.Connect,
				},
			},
		},
		{
			name: "all with grant",
			in:   "foo=C*T*c*",
			out:  "foo=C*T*c*",
			want: acl.Database{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Create | acl.Temporary | acl.Connect,
					GrantOptions: acl.Create | acl.Temporary | acl.Connect,
				},
			},
		},
		{
			name: "all with grant by role",
			in:   "foo=C*T*c*/bar",
			out:  "foo=C*T*c*/bar",
			want: acl.Database{
				ACL: acl.ACL{
					Role:         "foo",
					GrantedBy:    "bar",
					Privileges:   acl.Create | acl.Temporary | acl.Connect,
					GrantOptions: acl.Create | acl.Temporary | acl.Connect,
				},
			},
		},
		{
			name: "public all",
			in:   "=c",
			out:  "=c",
			want: acl.Database{
				ACL: acl.ACL{
					Role:       "",
					Privileges: acl.Connect,
				},
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.Database{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.Database{},
			fail: true,
		},
	}

	for i, test := range tests {
		if test.name == "" {
			t.Fatalf("test %d needs a name", i)
		}

		t.Run(test.name, func(t *testing.T) {
			aclItem, err := acl.Parse(test.in)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse ACLItem %+q: %v", test.in, err)
			}

			if err == nil && test.fail {
				t.Fatalf("expected failure")
			}

			if test.fail && err != nil {
				return
			}

			got, err := acl.NewDatabase(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse database ACL %+q: %v", test.in, err)
			}

			if err == nil && test.fail {
				t.Fatalf("expected failure")
			}

			if test.fail && err != nil {
				return
			}

			if out := test.want.String(); out != test.out {
				t.Fatalf("want %+q got %+q", test.out, out)
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Fatalf("bad: expected %v to equal %v", test.want, got)
			}
		})
	}
}
