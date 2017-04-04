package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestForeignDataWrapperString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		want acl.ForeignDataWrapper
		fail bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
		},
		{
			name: "all without grant",
			in:   "foo=U",
			out:  "foo=U",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Usage,
				},
			},
		},
		{
			name: "all with grant",
			in:   "foo=U*",
			out:  "foo=U*",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Usage,
					GrantOptions: acl.Usage,
				},
			},
		},
		{
			name: "all with grant by role",
			in:   "foo=U*/bar",
			out:  "foo=U*/bar",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role:         "foo",
					GrantedBy:    "bar",
					Privileges:   acl.Usage,
					GrantOptions: acl.Usage,
				},
			},
		},
		{
			name: "all mixed grant1",
			in:   "foo=U*",
			out:  "foo=U*",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Usage,
					GrantOptions: acl.Usage,
				},
			},
		},
		{
			name: "all mixed grant2",
			in:   "foo=U",
			out:  "foo=U",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Usage,
				},
			},
		},
		{
			name: "public all",
			in:   "=U*",
			out:  "=U*",
			want: acl.ForeignDataWrapper{
				ACL: acl.ACL{
					Role:         "",
					Privileges:   acl.Usage,
					GrantOptions: acl.Usage,
				},
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.ForeignDataWrapper{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.ForeignDataWrapper{},
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

			got, err := acl.NewForeignDataWrapper(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse FDW ACL %+q: %v", test.in, err)
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
