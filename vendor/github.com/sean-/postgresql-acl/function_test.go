package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestFunctionString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		want acl.Function
		fail bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.Function{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
		},
		{
			name: "all without grant",
			in:   "foo=X",
			out:  "foo=X",
			want: acl.Function{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Execute,
				},
			},
		},
		{
			name: "all with grant",
			in:   "foo=X*",
			out:  "foo=X*",
			want: acl.Function{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Execute,
					GrantOptions: acl.Execute,
				},
			},
		},
		{
			name: "all with grant by role",
			in:   "foo=X*/bar",
			out:  "foo=X*/bar",
			want: acl.Function{
				ACL: acl.ACL{
					Role:         "foo",
					GrantedBy:    "bar",
					Privileges:   acl.Execute,
					GrantOptions: acl.Execute,
				},
			},
		},
		{
			name: "all mixed grant1",
			in:   "foo=X*",
			out:  "foo=X*",
			want: acl.Function{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Execute,
					GrantOptions: acl.Execute,
				},
			},
		},
		{
			name: "all mixed grant2",
			in:   "foo=X",
			out:  "foo=X",
			want: acl.Function{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Execute,
				},
			},
		},
		{
			name: "public all",
			in:   "=X*",
			out:  "=X*",
			want: acl.Function{
				ACL: acl.ACL{
					Role:         "",
					Privileges:   acl.Execute,
					GrantOptions: acl.Execute,
				},
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.Function{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.Function{},
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

			got, err := acl.NewFunction(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse function ACL %+q: %v", test.in, err)
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
