package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestSequenceString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		want acl.Sequence
		fail bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.Sequence{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
		},
		{
			name: "all without grant",
			in:   "foo=rwU",
			out:  "foo=rwU",
			want: acl.Sequence{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Select | acl.Update | acl.Usage,
				},
			},
		},
		{
			name: "all with grant",
			in:   "foo=r*w*U*",
			out:  "foo=r*w*U*",
			want: acl.Sequence{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Select | acl.Update | acl.Usage,
					GrantOptions: acl.Select | acl.Update | acl.Usage,
				},
			},
		},
		{
			name: "all with grant and by",
			in:   "foo=r*w*U*/bar",
			out:  "foo=r*w*U*/bar",
			want: acl.Sequence{
				ACL: acl.ACL{
					Role:         "foo",
					GrantedBy:    "bar",
					Privileges:   acl.Select | acl.Update | acl.Usage,
					GrantOptions: acl.Select | acl.Update | acl.Usage,
				},
			},
		},
		{
			name: "public all",
			in:   "=rU",
			out:  "=rU",
			want: acl.Sequence{
				ACL: acl.ACL{
					Role:       "",
					Privileges: acl.Select | acl.Usage,
				},
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.Sequence{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.Sequence{},
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

			got, err := acl.NewSequence(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse sequence ACL %+q: %v", test.in, err)
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
