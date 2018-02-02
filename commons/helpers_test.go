package commons

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type _TestVars struct {
	AwsAccountId, Environment, AppName, Version, TeamName string
}

func TestIsIgnored(t *testing.T) {
	ignoreList := []string{"one", "two"}

	if IsIgnored("one", ignoreList) {
		t.Errorf("one should be ignored (%#v)", ignoreList)
	}

	if !IsIgnored("three", ignoreList) {
		t.Errorf("one should not be ignored (%#v)", ignoreList)
	}
}

func TestProcessTemplates_valid(t *testing.T) {
	testRemoveByGlob("test-fixtures/templates-valid/*.tf")
	defer testRemoveByGlob("test-fixtures/templates-valid/*.tf")

	vars := struct{ Yada string }{"yada"}
	files, err := ProcessTemplates("test-fixtures/templates-valid", "tpl", vars)
	if err != nil {
		t.Fatalf("Failed processing templates: %s", err.Error())
	}

	_, err = os.Stat("test-fixtures/templates-valid/one.tf")
	if err != nil {
		t.Fatalf("Expected one.tf to exist: %q", err.Error())
	}
	_, err = os.Stat("test-fixtures/templates-valid/two.tf")
	if err != nil {
		t.Fatalf("Expected two.tf to exist: %q", err.Error())
	}

	expectedFiles := []string{
		"test-fixtures/templates-valid/one.tf",
		"test-fixtures/templates-valid/two.tf",
	}
	if !reflect.DeepEqual(files, expectedFiles) {
		t.Fatalf("Unexpected list of files returned.\nExpected: %#v\nGiven:  %#v",
			expectedFiles, files)
	}
}

func TestProcessTemplates_invalid(t *testing.T) {
	testRemoveByGlob("test-fixtures/templates-invalid-syntax/*.tf")
	defer testRemoveByGlob("test-fixtures/templates-invalid-syntax/*.tf")

	vars := struct{ Yada string }{"yada"}
	_, err := ProcessTemplates("test-fixtures/templates-invalid-syntax", "tpl", vars)
	if err == nil {
		t.Fatalf("Expected error when processing invalid template.")
	}
	expectedSubset := "unexpected bad character"
	if !strings.Contains(err.Error(), expectedSubset) {
		t.Fatalf("Expected error to contain %q: %s", expectedSubset, err.Error())
	}

	testRemoveByGlob("test-fixtures/templates-missing-var/*.tf")
	defer testRemoveByGlob("test-fixtures/templates-missing-var/*.tf")

	_, err = ProcessTemplates("test-fixtures/templates-missing-var", "tpl", vars)
	if err == nil {
		t.Fatalf("Expected error when processing template w/ missing variable.")
	}
	expectedSubset = "can't evaluate field MissingName"
	if !strings.Contains(err.Error(), expectedSubset) {
		t.Fatalf("Expected error to contain %q: %s", expectedSubset, err.Error())
	}
}

func testRemoveByGlob(glob string) error {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, match := range matches {
		err := os.Remove(match)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestGenerateFromTemplate(t *testing.T) {
	outputPath := filepath.Join(os.TempDir(), "rt-template-output.tf")
	vars := _TestVars{
		AppName:  "test-app-name",
		TeamName: "Testing Team Name",
	}

	processTemplate("test-fixtures/template.tf.tpl", outputPath, vars)

	output, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Failed reading output path: %#v", err)
	}
	expectedOutput, err := ioutil.ReadFile("test-fixtures/template_processed.tf")
	if err != nil {
		t.Errorf("Failed reading fixture path: %#v", err)
	}

	if string(output) != string(expectedOutput) {
		t.Errorf("Generated files do not match.\nExpected: %#v\nGenerated: %#v",
			string(expectedOutput), string(output))
	}
}

func TestInferSlotIdFromVersionId(t *testing.T) {

	version := "123a8123"

	// With deployment id, should return deployment id
	expectedOutput := "SINGLE"
	output := InferSlotIdFromVersionId(expectedOutput, version)
	if output != expectedOutput {
		t.Errorf("Output: %q\nExpected output: %q", output, expectedOutput)
	}

	// With empty deployment id, should return version id
	output = InferSlotIdFromVersionId("", version)
	if output != version {
		t.Errorf("Output: %q\nExpected output: %q", output, version)
	}
}

func TestGetVersionPathOrSlotPathForApp(t *testing.T) {
	// Error if both version and deploy exist
	output, err := GetVersionPathOrSlotPathForApp("test-fixtures/directory", "one.txt", "two.txt")
	if err == nil {
		t.Errorf("Expected an error, none were thrown")
	}

	// slotPath if slot exists and version doesn't
	output, err = GetVersionPathOrSlotPathForApp("test-fixtures/directory", "one.txt", "nonexistant")
	if output != "test-fixtures/directory/one.txt" {
		t.Errorf("Output: %q\nExpected output: %q", output, "test-fixtures/directory/one.txt")
	}

	// versionPath if version exists and deploy doesn't
	output, err = GetVersionPathOrSlotPathForApp("test-fixtures/directory", "nonexistant", "two.txt")
	if output != "test-fixtures/directory/two.txt" {
		t.Errorf("Output: %q\nExpected output: %q", output, "test-fixtures/directory/two.txt")
	}

}

func TestDoesFileOrDirExist(t *testing.T) {

	// File should exist, output should come back as true
	output, err := doesFileOrDirExist("test-fixtures/directory")
	if err != nil {
		t.Fatalf("Error checking if file exists: %s", err.Error())
	}
	if !output {
		t.Errorf("Output: %q\nExpected output: %q", output, false)
	}

	// File shouldn't exist, output should come back as false
	output, err = doesFileOrDirExist("test-fixtures/something-nonexistant")
	if err != nil {
		t.Fatalf("Error checking if file exists: %s", err.Error())
	}
	if output {
		t.Errorf("Output: %q\nExpected output: %q", output, false)
	}

}

func TestReadAndDecodeHCLFromFile(t *testing.T) {
	tplVars := _TestVars{
		AwsAccountId: "123123123123",
		Environment:  "test",
		AppName:      "example",
		Version:      "asdf123",
	}

	rNonExistant, _ := os.Open("test-fixtures/something-nonexistant")
	_, err := DecodeHCLFromTemplate(rNonExistant, tplVars)
	if err == nil {
		t.Errorf("Expected an error as file does not exist")
	}

	rInvalid, _ := os.Open("test-fixtures/hcl/invalid.hcl.tpl")
	_, err = DecodeHCLFromTemplate(rInvalid, tplVars)
	if err == nil {
		t.Errorf("Expected an error as HCL is invalid")
	}

	rValid, _ := os.Open("test-fixtures/hcl/valid.hcl.tpl")
	output, _ := DecodeHCLFromTemplate(rValid, tplVars)
	expectedOutput := map[string]interface{}{
		"deployment_state": []map[string]interface{}{
			map[string]interface{}{
				"s3": []map[string]interface{}{
					map[string]interface{}{
						"key": "123123123123/example/asdf123/", "bucket": "ti-rt-deployment-state-test"},
				},
			},
		},
	}

	equal := reflect.DeepEqual(output, expectedOutput)
	if !equal {
		t.Errorf("Unexpected output, was expecting %v", expectedOutput)
	}
}
