package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestTablespaceString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		want acl.Tablespace
		fail bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
		},
		{
			name: "all without grant",
			in:   "foo=C",
			out:  "foo=C",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Create,
				},
			},
		},
		{
			name: "all with grant",
			in:   "foo=C*",
			out:  "foo=C*",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Create,
					GrantOptions: acl.Create,
				},
			},
		},
		{
			name: "all with grant by role",
			in:   "foo=C*/bar",
			out:  "foo=C*/bar",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role:         "foo",
					GrantedBy:    "bar",
					Privileges:   acl.Create,
					GrantOptions: acl.Create,
				},
			},
		},
		{
			name: "all mixed grant1",
			in:   "foo=C*",
			out:  "foo=C*",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Create,
					GrantOptions: acl.Create,
				},
			},
		},
		{
			name: "all mixed grant2",
			in:   "foo=C",
			out:  "foo=C",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Create,
				},
			},
		},
		{
			name: "public all",
			in:   "=C*",
			out:  "=C*",
			want: acl.Tablespace{
				ACL: acl.ACL{
					Role:         "",
					Privileges:   acl.Create,
					GrantOptions: acl.Create,
				},
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.Tablespace{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.Tablespace{},
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

			got, err := acl.NewTablespace(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse tablespace ACL %+q: %v", test.in, err)
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
