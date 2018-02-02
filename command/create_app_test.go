package command

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestParseVariablesConfig(t *testing.T) {
	config, err := parseVariablesConfig("test-fixtures/config.variables.json")
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}

	expected := map[string]interface{}{
		"Servers": map[string]interface{}{
			"description": "The number of servers",
			"default":     float64(3),
		},
	}
	if !reflect.DeepEqual(config, expected) {
		t.Fatalf("Generated:\n%#v\nexpected:\n%#v", config, expected)
	}
}

func TestAskForVariables_basic(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	variables := map[string]interface{}{
		"region": map[string]interface{}{
			"default":     "us-west-2",
			"description": "AWS Region",
		},
		"servers": map[string]interface{}{
			"description": "Number of servers",
		},
	}

	go w.Write([]byte("yada\n3\n"))
	vars, err := askForVariables(r, output, variables)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}

	expectedOutput := "\nAWS Region (us-west-2): \nNumber of servers: "
	if output.String() != expectedOutput {
		t.Errorf("Output: %q\nExpected output: %q", output, expectedOutput)
	}

	expectedVars := map[string]string{
		"servers": "3",
		"region":  "yada",
	}
	if !reflect.DeepEqual(vars, expectedVars) {
		t.Fatalf("Variables: %q\nExpected variables: %q", vars, expectedVars)
	}
}

func TestAskForVariables_useDefault(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	variables := map[string]interface{}{
		"region": map[string]interface{}{
			"default":     "us-west-2",
			"description": "AWS Region",
		},
		"servers": map[string]interface{}{
			"description": "Number of servers",
		},
	}

	go w.Write([]byte("\n3\n"))
	vars, err := askForVariables(r, output, variables)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}

	expectedOutput := "\nAWS Region (us-west-2): \nNumber of servers: "
	if output.String() != expectedOutput {
		t.Errorf("Output: %q\nExpected output: %q", output, expectedOutput)
	}

	expectedVars := map[string]string{
		"servers": "3",
		"region":  "us-west-2",
	}
	if !reflect.DeepEqual(vars, expectedVars) {
		t.Fatalf("Variables: %q\nExpected variables: %q", vars, expectedVars)
	}
}

func TestAskForVariables_required(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	variables := map[string]interface{}{
		"region": map[string]interface{}{
			"default":     "us-west-2",
			"description": "AWS Region",
		},
		"servers": map[string]interface{}{
			"description": "Number of servers",
		},
	}

	go w.Write([]byte("\n\n0\n"))
	vars, err := askForVariables(r, output, variables)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}

	expectedOutput := "\nAWS Region (us-west-2): \nNumber of servers: \nNumber of servers: "
	if output.String() != expectedOutput {
		t.Errorf("Output: %q\nExpected output: %q", output, expectedOutput)
	}

	expectedVars := map[string]string{
		"servers": "0",
		"region":  "us-west-2",
	}
	if !reflect.DeepEqual(vars, expectedVars) {
		t.Fatalf("Variables: %q\nExpected variables: %q", vars, expectedVars)
	}
}

func TestCopyTree(t *testing.T) {
	targetPath := "test-fixtures/target-path"
	os.RemoveAll(targetPath)
	defer os.RemoveAll(targetPath)

	err := copyTree("test-fixtures/test-skeleton", targetPath)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}

	// TODO: Compare dir trees
}
