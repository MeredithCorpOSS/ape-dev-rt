package commons

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl"
)

func IsIgnored(filePath string, staticFilePaths []string) bool {
	for _, path := range staticFilePaths {
		if path == filePath {
			return false
		}
	}

	return true
}

func ProcessTemplates(path, fileSuffix string, vars interface{}) ([]string, error) {
	return processTemplates(path, fileSuffix, vars, false)
}

func ProcessTemplatesAndDelete(path, fileSuffix string, vars interface{}) ([]string, error) {
	return processTemplates(path, fileSuffix, vars, true)
}

func processTemplates(path, fileSuffix string, vars interface{}, deleteTpl bool) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(path, "*."+fileSuffix))
	if err != nil {
		return nil, err
	}

	targetFiles := make([]string, 0)
	for _, match := range matches {
		log.Printf("[DEBUG] Processing %q", match)
		targetFilename := strings.TrimSuffix(match, "."+fileSuffix)

		err := processTemplate(match, targetFilename, vars)
		if err != nil {
			return nil, err
		}
		targetFiles = append(targetFiles, targetFilename)
		if deleteTpl {
			log.Printf("[DEBUG] Deleting processed template: %q", match)
			err = os.Remove(match)
			if err != nil {
				return nil, err
			}
		}
	}

	return targetFiles, nil
}

func processTemplate(sourcePath, targetPath string, vars interface{}) error {
	log.Printf("Reading a template from %s", sourcePath)
	bytes, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	funcMap := map[string]interface{}{"mkSlice": mkSlice}
	t, err := template.New(sourcePath).Funcs(template.FuncMap(funcMap)).Parse(string(bytes))
	if err != nil {
		return err
	}

	// Generate output to target path
	log.Printf("Creating target path at %s", targetPath)
	f, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("Creation of target path failed: %s", err)
	}
	w := bufio.NewWriter(f)

	err = t.Execute(w, vars)
	if err != nil {
		return fmt.Errorf("Generating template failed: %s", err)
	}
	w.Flush()
	log.Printf("Written output into %s", targetPath)

	return nil
}

func mkSlice(args ...interface{}) []interface{} {
    return args
}

func InferSlotIdFromVersionId(slotId, version string) string {
	if slotId == "" || slotId == "master" {
		return version
	}
	return slotId
}

func GetVersionPathOrSlotPathForApp(appPath, slotSuffix, versionSuffix string) (string, error) {
	slotPath := path.Join(appPath, slotSuffix)
	slotPathExists, err := doesFileOrDirExist(slotPath)
	if err != nil {
		return "", fmt.Errorf("Checking existence of %s failed: %s", slotPath, err)
	}

	versionPath := path.Join(appPath, versionSuffix)
	versionPathExists, err := doesFileOrDirExist(versionPath)
	if err != nil {
		return "", fmt.Errorf("Checking existence of %s failed: %s", versionPath, err)
	}

	if slotPathExists && versionPathExists {
		return "", fmt.Errorf("Expected either slot (%s) or version (%s) path to exist, both exist.\nThis may require manual intervention in S3 to figure out whether the app was actually migrated.", slotPath, versionPath)
	}

	if slotPathExists {
		return slotPath, nil
	}

	return versionPath, nil
}

func doesFileOrDirExist(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func DecodeHCLFromTemplate(templateReader io.Reader, variables interface{}) (map[string]interface{}, error) {
	b := bytes.Buffer{}
	_, err := b.ReadFrom(templateReader)
	if err != nil {
		return nil, err
	}
	hclTemplateInBytes := b.Bytes()

	// Parse the template
	t := template.Must(template.New("-").Parse(string(hclTemplateInBytes)))

	var processedTplString bytes.Buffer
	err = t.Execute(&processedTplString, variables)
	if err != nil {
		return nil, err
	}

	// And now let's decode the HCL
	var configs map[string]interface{}
	err = hcl.Decode(&configs, processedTplString.String())
	if err != nil {
		return nil, fmt.Errorf("Unable to decode HCL: %s", err)
	}

	return configs, nil
}
