package acl_test

import (
	"reflect"
	"testing"

	acl "github.com/sean-/postgresql-acl"
)

func TestSchemaString(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		out     string
		want    acl.Schema
		grants  []string
		revokes []string
		fail    bool
	}{
		{
			name: "default",
			in:   "foo=",
			out:  "foo=",
			want: acl.Schema{
				ACL: acl.ACL{
					Role: "foo",
				},
			},
			grants:  []string{},
			revokes: []string{},
		},
		{
			name: "all without grant",
			in:   "foo=UC",
			out:  "foo=UC",
			want: acl.Schema{
				ACL: acl.ACL{
					Role:       "foo",
					Privileges: acl.Create | acl.Usage,
				},
			},
			grants: []string{
				`GRANT CREATE ON SCHEMA "all without grant" TO "foo"`,
				`GRANT USAGE ON SCHEMA "all without grant" TO "foo"`,
			},
			revokes: []string{
				`REVOKE CREATE ON SCHEMA "all without grant" FROM "foo"`,
				`REVOKE USAGE ON SCHEMA "all without grant" FROM "foo"`,
			},
		},
		{
			name: "all with grant",
			in:   "foo=U*C*",
			out:  "foo=U*C*",
			want: acl.Schema{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Create | acl.Usage,
					GrantOptions: acl.Create | acl.Usage,
				},
			},
			grants: []string{
				`GRANT CREATE ON SCHEMA "all with grant" TO "foo" WITH GRANT OPTION`,
				`GRANT USAGE ON SCHEMA "all with grant" TO "foo" WITH GRANT OPTION`,
			},
			revokes: []string{
				`REVOKE GRANT OPTION FOR CREATE ON SCHEMA "all with grant" FROM "foo"`,
				`REVOKE GRANT OPTION FOR USAGE ON SCHEMA "all with grant" FROM "foo"`,
			},
		},
		{
			name: "all with grant by role",
			in:   "foo=U*C*/bar",
			out:  "foo=U*C*/bar",
			want: acl.Schema{
				ACL: acl.ACL{
					Role:         "foo",
					GrantedBy:    "bar",
					Privileges:   acl.Create | acl.Usage,
					GrantOptions: acl.Create | acl.Usage,
				},
			},
			grants: []string{
				`GRANT CREATE ON SCHEMA "all with grant by role" TO "foo" WITH GRANT OPTION`,
				`GRANT USAGE ON SCHEMA "all with grant by role" TO "foo" WITH GRANT OPTION`,
			},
			revokes: []string{
				`REVOKE GRANT OPTION FOR CREATE ON SCHEMA "all with grant by role" FROM "foo"`,
				`REVOKE GRANT OPTION FOR USAGE ON SCHEMA "all with grant by role" FROM "foo"`,
			},
		},
		{
			name: "all mixed grant1",
			in:   "foo=U*C",
			out:  "foo=U*C",
			want: acl.Schema{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Create | acl.Usage,
					GrantOptions: acl.Usage,
				},
			},
			grants: []string{
				`GRANT CREATE ON SCHEMA "all mixed grant1" TO "foo"`,
				`GRANT USAGE ON SCHEMA "all mixed grant1" TO "foo" WITH GRANT OPTION`,
			},
			revokes: []string{
				`REVOKE CREATE ON SCHEMA "all mixed grant1" FROM "foo"`,
				`REVOKE GRANT OPTION FOR USAGE ON SCHEMA "all mixed grant1" FROM "foo"`,
			},
		},
		{
			name: "all mixed grant2",
			in:   "foo=UC*",
			out:  "foo=UC*",
			want: acl.Schema{
				ACL: acl.ACL{
					Role:         "foo",
					Privileges:   acl.Create | acl.Usage,
					GrantOptions: acl.Create,
				},
			},
			grants: []string{
				`GRANT CREATE ON SCHEMA "all mixed grant2" TO "foo" WITH GRANT OPTION`,
				`GRANT USAGE ON SCHEMA "all mixed grant2" TO "foo"`,
			},
			revokes: []string{
				`REVOKE GRANT OPTION FOR CREATE ON SCHEMA "all mixed grant2" FROM "foo"`,
				`REVOKE USAGE ON SCHEMA "all mixed grant2" FROM "foo"`,
			},
		},
		{
			name: "public all",
			in:   "=U*C*",
			out:  "=U*C*",
			want: acl.Schema{
				ACL: acl.ACL{
					Role:         "",
					Privileges:   acl.Create | acl.Usage,
					GrantOptions: acl.Create | acl.Usage,
				},
			},
			grants: []string{
				`GRANT CREATE ON SCHEMA "public all" TO PUBLIC WITH GRANT OPTION`,
				`GRANT USAGE ON SCHEMA "public all" TO PUBLIC WITH GRANT OPTION`,
			},
			revokes: []string{
				`REVOKE GRANT OPTION FOR CREATE ON SCHEMA "public all" FROM PUBLIC`,
				`REVOKE GRANT OPTION FOR USAGE ON SCHEMA "public all" FROM PUBLIC`,
			},
		},
		{
			name: "invalid input1",
			in:   "bar*",
			want: acl.Schema{},
			fail: true,
		},
		{
			name: "invalid input2",
			in:   "%",
			want: acl.Schema{},
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

			got, err := acl.NewSchema(aclItem)
			if err != nil && !test.fail {
				t.Fatalf("unable to parse schema ACL %+q: %v", test.in, err)
			}

			if out := test.want.String(); out != test.out {
				t.Fatalf("want %+q got %+q", test.out, out)
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Fatalf("bad: expected %v to equal %v", test.want, got)
			}

			grants := got.Grants(test.name)
			if !reflect.DeepEqual(test.grants, grants) {
				t.Fatalf("bad: expected %#v to equal %#v", test.grants, grants)
			}

			revokes := got.Revokes(test.name)
			if !reflect.DeepEqual(test.revokes, revokes) {
				t.Fatalf("bad: expected %v to equal %v", test.revokes, revokes)
			}
		})
	}
}
