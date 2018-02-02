package terraform

import (
	"reflect"
	"strings"
	"testing"
)

func TestApplyGraphBuilder_impl(t *testing.T) {
	var _ GraphBuilder = new(ApplyGraphBuilder)
}

func TestApplyGraphBuilder(t *testing.T) {
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: []string{"root"},
				Resources: map[string]*InstanceDiff{
					// Verify noop doesn't show up in graph
					"aws_instance.noop": &InstanceDiff{},

					"aws_instance.create": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},

					"aws_instance.other": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},
				},
			},

			&ModuleDiff{
				Path: []string{"root", "child"},
				Resources: map[string]*InstanceDiff{
					"aws_instance.create": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},

					"aws_instance.other": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},
				},
			},
		},
	}

	b := &ApplyGraphBuilder{
		Module:        testModule(t, "graph-builder-apply-basic"),
		Diff:          diff,
		Providers:     []string{"aws"},
		Provisioners:  []string{"exec"},
		DisableReduce: true,
	}

	g, err := b.Build(RootModulePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(g.Path, RootModulePath) {
		t.Fatalf("bad: %#v", g.Path)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testApplyGraphBuilderStr)
	if actual != expected {
		t.Fatalf("bad: %s", actual)
	}
}

// This tests the ordering of two resources that are both CBD that
// require destroy/create.
func TestApplyGraphBuilder_doubleCBD(t *testing.T) {
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: []string{"root"},
				Resources: map[string]*InstanceDiff{
					"aws_instance.A": &InstanceDiff{
						Destroy: true,
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},

					"aws_instance.B": &InstanceDiff{
						Destroy: true,
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},
				},
			},
		},
	}

	b := &ApplyGraphBuilder{
		Module:        testModule(t, "graph-builder-apply-double-cbd"),
		Diff:          diff,
		Providers:     []string{"aws"},
		Provisioners:  []string{"exec"},
		DisableReduce: true,
	}

	g, err := b.Build(RootModulePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(g.Path, RootModulePath) {
		t.Fatalf("bad: %#v", g.Path)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testApplyGraphBuilderDoubleCBDStr)
	if actual != expected {
		t.Fatalf("bad: %s", actual)
	}
}

// This tests the ordering of destroying a single count of a resource.
func TestApplyGraphBuilder_destroyCount(t *testing.T) {
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: []string{"root"},
				Resources: map[string]*InstanceDiff{
					"aws_instance.A.1": &InstanceDiff{
						Destroy: true,
					},

					"aws_instance.B": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"name": &ResourceAttrDiff{
								Old: "",
								New: "foo",
							},
						},
					},
				},
			},
		},
	}

	b := &ApplyGraphBuilder{
		Module:        testModule(t, "graph-builder-apply-count"),
		Diff:          diff,
		Providers:     []string{"aws"},
		Provisioners:  []string{"exec"},
		DisableReduce: true,
	}

	g, err := b.Build(RootModulePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(g.Path, RootModulePath) {
		t.Fatalf("bad: %#v", g.Path)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testApplyGraphBuilderDestroyCountStr)
	if actual != expected {
		t.Fatalf("bad: %s", actual)
	}
}

const testApplyGraphBuilderStr = `
aws_instance.create
  provider.aws
aws_instance.other
  aws_instance.create
  provider.aws
meta.count-boundary (count boundary fixup)
  aws_instance.create
  aws_instance.other
  module.child.aws_instance.create
  module.child.aws_instance.other
  module.child.provider.aws
  module.child.provisioner.exec
  provider.aws
module.child.aws_instance.create
  module.child.provider.aws
  module.child.provisioner.exec
module.child.aws_instance.other
  module.child.aws_instance.create
  module.child.provider.aws
module.child.provider.aws
  provider.aws
module.child.provisioner.exec
provider.aws
`

const testApplyGraphBuilderDoubleCBDStr = `
aws_instance.A
  provider.aws
aws_instance.A (destroy)
  aws_instance.A
  aws_instance.B
  aws_instance.B (destroy)
  provider.aws
aws_instance.B
  aws_instance.A
  provider.aws
aws_instance.B (destroy)
  aws_instance.B
  provider.aws
meta.count-boundary (count boundary fixup)
  aws_instance.A
  aws_instance.A (destroy)
  aws_instance.B
  aws_instance.B (destroy)
  provider.aws
provider.aws
`

const testApplyGraphBuilderDestroyCountStr = `
aws_instance.A[1] (destroy)
  provider.aws
aws_instance.B
  aws_instance.A[1] (destroy)
  provider.aws
meta.count-boundary (count boundary fixup)
  aws_instance.A[1] (destroy)
  aws_instance.B
  provider.aws
provider.aws
`
