package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestTableString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		want acl.Table
		fail bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.Table{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
		},
		{
			name: "all without grant",
			in:   "foo=arwdDxt",
			out:  "foo=arwdDxt",
			want: acl.Table{
				ACL: acl.ACL{
					Role: "foo",
					Privileges: acl.Delete |
						acl.Insert |
						acl.References |
						acl.Select |
						acl.Trigger |
						acl.Truncate |
						acl.Update,
				},
			},
		},
		{
			name: "all with grant",
			in:   "foo=a*r*w*d*D*x*t*",
			out:  "foo=a*r*w*d*D*x*t*",
			want: acl.Table{
				ACL: acl.ACL{
					Role: "foo",
					Privileges: acl.Delete |
						acl.Insert |
						acl.References |
						acl.Select |
						acl.Trigger |
						acl.Truncate |
						acl.Update,
					GrantOptions: acl.Delete |
						acl.Insert |
						acl.References |
						acl.Select |
						acl.Trigger |
						acl.Truncate |
						acl.Update,
				},
			},
		},
		{
			name: "all with grant and by",
			in:   "foo=a*r*w*d*D*x*t*/bar",
			out:  "foo=a*r*w*d*D*x*t*/bar",
			want: acl.Table{
				ACL: acl.ACL{
					Role:      "foo",
					GrantedBy: "bar",
					Privileges: acl.Delete |
						acl.Insert |
						acl.References |
						acl.Select |
						acl.Trigger |
						acl.Truncate |
						acl.Update,
					GrantOptions: acl.Delete |
						acl.Insert |
						acl.References |
						acl.Select |
						acl.Trigger |
						acl.Truncate |
						acl.Update,
				},
			},
		},
		{
			name: "public all",
			in:   "=r",
			out:  "=r",
			want: acl.Table{
				ACL: acl.ACL{
					Role:       "",
					Privileges: acl.Select,
				},
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.Table{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.Table{},
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

			got, err := acl.NewTable(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse table ACL %+q: %v", test.in, err)
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
