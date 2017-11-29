package terraform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestPlan(t *testing.T) {
	cfg := `
variable "name" {}
variable "profile" {}
provider "aws" {
  profile = "${var.profile}"
  region = "us-west-2"
}
resource "aws_sns_topic" "s" {
  name = "tf-test-${var.name}"
}`
	tmpDir, profileName := createTempTerraformEnv(cfg, t)
	defer os.RemoveAll(tmpDir)
	planFilePath := path.Join(tmpDir, "planfile")

	input := PlanInput{
		PlanFilePath: planFilePath,
		Refresh:      false,
		RootPath:     tmpDir,
		Variables: map[string]string{
			"name":    "umarsticks",
			"profile": profileName,
		},
	}
	out, err := Plan(&input)
	if err != nil {
		t.Fatal(err)
	}
	expectedOut := &PlanOutput{
		ExitCode: 0,
		Stdout: fmt.Sprintf(`The Terraform execution plan has been generated and is shown below.
Resources are shown in alphabetical order for quick scanning. Green resources
will be created (or destroyed and then created if an existing resource
exists), yellow resources are being changed in-place, and red resources
will be destroyed. Cyan entries are data sources to be read.

Your plan was also saved to the path below. Call the "apply" subcommand
with this plan file and Terraform will exactly execute this execution
plan.

Path: %s

[32m+ aws_sns_topic.s
[0m    arn:    "<computed>"
    name:   "tf-test-umarsticks"
    policy: "<computed>"
[0m
[0m
[0m[1mPlan:[0m 1 to add, 0 to change, 0 to destroy.[0m
`, planFilePath),
		Stderr:   "",
		Warnings: []string{},
		Diff: &PlanResourceDiff{
			ToCreate: 1,
			ToChange: 0,
			ToRemove: 0,
		},
	}
	if !reflect.DeepEqual(*out, *expectedOut) {
		t.Fatalf("Plan output doesn't match.\nGiven:    %#v\nExpected: %#v\n",
			*out, *expectedOut)
	}
}

func TestApplyDestroy(t *testing.T) {
	cfg := `
variable "name" {}
variable "profile" {}
provider "aws" {
  profile = "${var.profile}"
  region = "us-west-2"
}
resource "aws_sns_topic" "s" {
  name = "tf-test-${var.name}"
}
output "yololo" {
  value = "yada"
}
`
	tmpDir, profileName := createTempTerraformEnv(cfg, t)
	defer os.RemoveAll(tmpDir)
	planFilePath := path.Join(tmpDir, "planfile")
	vars := map[string]string{
		"name":    "umarsticks",
		"profile": profileName,
	}

	input := PlanInput{
		PlanFilePath: planFilePath,
		Refresh:      false,
		RootPath:     tmpDir,
		Variables:    vars,
	}
	_, err := Plan(&input)
	if err != nil {
		t.Fatal(err)
	}

	applyInput := ApplyInput{
		PlanFilePath: planFilePath,
		Refresh:      false,
		RootPath:     tmpDir,
	}

	applyOut, err := Apply(&applyInput)
	expectedOut := &ApplyOutput{
		Stdout: `[0m[1maws_sns_topic.s: Creating...[21m
  arn:    "" => "<computed>"
  name:   "" => "tf-test-umarsticks"
  policy: "" => "<computed>"[0m
[0m[1maws_sns_topic.s: Creation complete[21m[0m
[0m[1m[32m
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.[0m
[0m
The state of your infrastructure has been saved to the path
below. This state is required to modify and destroy your
infrastructure, so keep it safe. To inspect the complete state
use the ` + "`terraform show`" + ` command.

State path: terraform.tfstate[0m
[0m[1m[32m
Outputs:

yololo = yada[0m
`,
		Stderr: "",
		Diff: &ResourceDiff{
			Created: 1,
			Changed: 0,
			Removed: 0,
		},
		ExitCode: 0,
		Outputs: map[string]string{
			"yololo": "yada",
		},
		Warnings: []string{},
	}
	if !reflect.DeepEqual(*applyOut, *expectedOut) {
		t.Fatalf("Apply output doesn't match. There may be DANGLING RESOURCES.\nGiven:    %#v\nExpected: %#v\n",
			*applyOut, *expectedOut)
	}

	destroyOut, err := Destroy(&DestroyInput{
		RootPath:  tmpDir,
		Variables: vars,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedDestroyOut := &DestroyOutput{
		ExitCode: 0,
		Stdout: `[0m[1maws_sns_topic.s: Destroying...[21m[0m
[0m[1maws_sns_topic.s: Destruction complete[21m[0m
[0m[1m[32m
Destroy complete! Resources: 1 destroyed.[0m
`,
		Stderr:   "",
		Warnings: []string{},
		Diff: &ResourceDiff{
			Created: 0,
			Changed: 0,
			Removed: 1,
		},
	}
	if !reflect.DeepEqual(*destroyOut, *expectedDestroyOut) {
		t.Fatalf("Destroy output doesn't match. There may be DANGLING RESOURCES.\nGiven:    %#v\nExpected: %#v\n",
			*destroyOut, *expectedDestroyOut)
	}
}

func TestOutput(t *testing.T) {
	cfg := `
variable "one" {}
variable "two" {}
output "alpha" {
  value = "one-${var.one},two-${var.two}"
}
output "static" {
  value = "yololo"
}
`
	tmpDir, _ := createTempTerraformEnv(cfg, t)
	defer os.RemoveAll(tmpDir)
	vars := map[string]string{
		"one": "aaa",
		"two": "bbb",
	}

	_, err := Apply(&ApplyInput{
		Variables: vars,
		Refresh:   false,
		RootPath:  tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	outputs, err := Output(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutputs := map[string]string{
		"alpha":  "one-aaa,two-bbb",
		"static": "yololo",
	}
	if !reflect.DeepEqual(outputs, expectedOutputs) {
		t.Fatalf("Unexpected outputs.\nGiven:    %#v\nExpected: %#v\n", outputs, expectedOutputs)
	}
}

func TestShow(t *testing.T) {
	cfg := `
variable "single" {}
`
	tmpDir, _ := createTempTerraformEnv(cfg, t)
	fmt.Println("TMPTMP: ", tmpDir)
	defer os.RemoveAll(tmpDir)
	vars := map[string]string{
		"single": "aaa",
	}

	_, err := Apply(&ApplyInput{
		Variables: vars,
		Refresh:   false,
		RootPath:  tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	output, err := Show(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := ""
	if output != expectedOutput {
		t.Fatalf("Unexpected output.\nGiven:    %#v\nExpected: %#v\n", output, expectedOutput)
	}
}

func TestGet(t *testing.T) {
	// TODO
}

func createTempTerraformEnv(content string, t *testing.T) (string, string) {
	profileName := os.Getenv("RT_ACC_AWS_PROFILE")
	if profileName == "" {
		t.Skip("Please set AWS profile name (RT_ACC_AWS_PROFILE)")
	}

	tmpDir := path.Join(os.TempDir(), "tf-test")
	err := os.Mkdir(tmpDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path.Join(tmpDir, "test.tf"))
	if err != nil {
		t.Fatalf("Failed creating temp file: %s", err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatalf("Failed saving temp config: %s", err)
	}
	return tmpDir, profileName
}

func TestParseSumsFromPlanOutput_add(t *testing.T) {
	testOutput := `yada yada yada o834u984 (*£@&*£@)*

    Plan: 1 to add, 0 to change, 0 to destroy.
`

	expectedDiff := &PlanResourceDiff{
		ToCreate: 1,
		ToChange: 0,
		ToRemove: 0,
	}

	actualDiff, err := parseDiffFromPlanOutput(testOutput)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}
	if !reflect.DeepEqual(expectedDiff, actualDiff) {
		t.Fatalf("Sums don't match. Expected: %#v\nGiven: %#v",
			expectedDiff, actualDiff)
	}
}

func TestParseSumsFromPlanOutput_change(t *testing.T) {
	testOutput := `yada yada yada o834u984 (*£@&*£@)*

    Plan: 0 to add, 45 to change, 0 to destroy.
`

	expectedDiff := &PlanResourceDiff{
		ToCreate: 0,
		ToChange: 45,
		ToRemove: 0,
	}

	actualDiff, err := parseDiffFromPlanOutput(testOutput)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}
	if !reflect.DeepEqual(expectedDiff, actualDiff) {
		t.Fatalf("Sums don't match. Expected: %#v\nGiven: %#v",
			expectedDiff, actualDiff)
	}
}

func TestParseSumsFromPlanOutput_destroy(t *testing.T) {
	testOutput := `yada yada yada o834u984 (*£@&*£@)*

    Plan: 0 to add, 0 to change, 8 to destroy.
`

	expectedDiff := &PlanResourceDiff{
		ToCreate: 0,
		ToChange: 0,
		ToRemove: 8,
	}

	actualDiff, err := parseDiffFromPlanOutput(testOutput)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}
	if !reflect.DeepEqual(expectedDiff, actualDiff) {
		t.Fatalf("Sums don't match. Expected: %#v\nGiven: %#v",
			expectedDiff, actualDiff)
	}
}

func TestParseSumsFromPlanOutput_combined(t *testing.T) {
	testOutput := `yada yada yada o834u984 (*£@&*£@)*

    Plan: 11 to add, 23 to change, 9 to destroy.
`

	expectedDiff := &PlanResourceDiff{
		ToCreate: 11,
		ToChange: 23,
		ToRemove: 9,
	}

	actualDiff, err := parseDiffFromPlanOutput(testOutput)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}
	if !reflect.DeepEqual(expectedDiff, actualDiff) {
		t.Fatalf("Sums don't match. Expected: %#v\nGiven: %#v",
			expectedDiff, actualDiff)
	}
}

func TestParseSumsFromPlanOutput_noChanges(t *testing.T) {
	testOutput := `No changes. Infrastructure is up-to-date. This means that Terraform
could not detect any differences between your configuration and
the real physical resources that exist. As a result, Terraform
doesn't need to do anything.
`

	actualDiff, err := parseDiffFromPlanOutput(testOutput)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}
	if actualDiff.ToChange != 0 {
		t.Fatalf("Unexpected ToChange: %d, expected zero", actualDiff.ToChange)
	}
	if actualDiff.ToCreate != 0 {
		t.Fatalf("Unexpected ToCreate: %d, expected zero", actualDiff.ToCreate)
	}
	if actualDiff.ToRemove != 0 {
		t.Fatalf("Unexpected ToRemove: %d, expected zero", actualDiff.ToRemove)
	}
}

func TestParseOutWarnings(t *testing.T) {
	cases := []struct {
		Output           string
		ExpectedWarnings []string
	}{
		0: {
			Output: `There are warnings and/or errors related to your configuration. Please
fix these before continuing.

[33mWarnings:
[0m[0m
[33m  * template_file.blah: using template_file as a resource is deprecated; consider using the data source instead[0m[0m
[33m  * template_file.current: using template_file as a resource is deprecated; consider using the data source instead[0m[0m

`,
			ExpectedWarnings: []string{
				"template_file.blah: using template_file as a resource is deprecated; consider using the data source instead",
				"template_file.current: using template_file as a resource is deprecated; consider using the data source instead",
			},
		},
		1: {
			Output:           "The Terraform execution plan has been generated and is shown below.\nResources are shown in alphabetical order for quick scanning. Green resources\nwill be created (or destroyed and then created if an existing resource\nexists), yellow resources are being changed in-place, and red resources\nwill be destroyed.\n\nYour plan was also saved to the path below. Call the \"apply\" subcommand\nwith this plan file and Terraform will exactly execute this execution\nplan.\n\nPath: /var/folders/dj/8v7ccb5n3838_mkbkfrxn59jr0_589/T/planfile\n\n\x1b[32m+ aws_sns_topic.s\n\x1b[0m    arn:    \"\" => \"<computed>\"\n    name:   \"\" => \"tf-test-umarsticks\"\n    policy: \"\" => \"<computed>\"\n\x1b[0m\n\x1b[0m\n\x1b[0m\x1b[1mPlan:\x1b[0m 1 to add, 0 to change, 0 to destroy.\x1b[0m\n",
			ExpectedWarnings: []string{},
		},
	}

	for i, c := range cases {
		warnings := parseOutWarnings(c.Output)
		if !reflect.DeepEqual(warnings, c.ExpectedWarnings) {
			t.Fatalf("Case %d: Unexpected warnings parsed out.\nExpected: %#v\nGiven:    %#v\n",
				i, c.ExpectedWarnings, warnings)
		}
	}
}

func TestParseOutputsFromApplyOutput(t *testing.T) {
	cases := []struct {
		Input           string
		ExpectedOutputs map[string]string
	}{
		0: {
			Input: `
State path: terraform.tfstate[0m
[0m[1m[32m
Outputs:

id = vpc-6cc7f308
yada = yololo[0m
`,
			ExpectedOutputs: map[string]string{
				"id":   "vpc-6cc7f308",
				"yada": "yololo",
			},
		},
		1: {
			Input: `State path: terraform.tfstate[0m
[0m[1m[32m
Outputs:

yololo = yada[0m
`,
			ExpectedOutputs: map[string]string{
				"yololo": "yada",
			},
		},
		2: {
			Input: `Outputs:

mylist = [
    us-east-1a,
    us-east-1c,
    us-east-1d,
    us-east-1e
]
mymap = {
  one = 111
  two = 222
}
mystring = sdfasdfasdf[0m
`,
			ExpectedOutputs: map[string]string{
				"mystring": "sdfasdfasdf",
			},
		},
	}

	for i, c := range cases {
		outputs := parseOutputsFromApplyOutput(c.Input)
		if !reflect.DeepEqual(outputs, c.ExpectedOutputs) {
			t.Fatalf("Case %d: Unexpected outputs parsed out.\nExpected: %#v\nGiven:    %#v\n",
				i, c.ExpectedOutputs, outputs)
		}
	}
}

func TestParseDiffFromApplyOutput(t *testing.T) {
	cases := []struct {
		Input        string
		ExpectedDiff *ResourceDiff
	}{
		0: {
			Input: `[0m[1maws_sns_topic.s: Destroying...[21m[0m
[0m[1maws_sns_topic.s: Destruction complete[21m[0m
[0m[1m[32m
Apply complete! Resources: 0 added, 0 changed, 1 destroyed.[0m
`,
			ExpectedDiff: &ResourceDiff{Created: 0, Changed: 0, Removed: 1},
		},
		1: {
			Input: `[0m[1maws_sns_topic.s: Destroying...[21m[0m
[0m[1maws_sns_topic.s: Destruction complete[21m[0m
[0m[1m[32m
Destroy complete! Resources: 1 destroyed.[0m
`,
			ExpectedDiff: &ResourceDiff{Created: 0, Changed: 0, Removed: 1},
		},
	}

	for i, c := range cases {
		diff, err := parseDiffFromApplyOutput(c.Input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(*diff, *c.ExpectedDiff) {
			t.Fatalf("Case %d: Unexpected diff parsed out.\nExpected: %#v\nGiven:    %#v\n",
				i, *c.ExpectedDiff, *diff)
		}
	}
}

func TestGenerateBackendConfig(t *testing.T) {
	var test_config map[string]string
	test_config = make(map[string]string)
	test_config["key"] = "some_path"
	var test_state = RemoteState{"s3", test_config}
	var filename = "backend-config.tf.json"
	_, err := GenerateBackendConfig(&test_state, "./")
	if err != nil {
		t.Fatalf("GenerateBackendConfig failed %s", err)
	}

	raw, err := ioutil.ReadFile(path.Join(".", filename))
	if err != nil {
		fmt.Println(err.Error())
		t.Fatalf("Failed to open json file")
	}
	var f interface{}
	err = json.Unmarshal(raw, &f)
	if err != nil {
		fmt.Println(err.Error())
		t.Fatalf("Failed to read json")
	}

	m := f.(map[string]interface{})
	_, ok := m["terraform"].([]interface{})
	if ok != true {
		t.Fatalf("Incorrect format backend config")
	}

	err = os.Remove(path.Join(".", filename))
	if err != nil {
		fmt.Println("failed to remove test backend config file")
	}
}
