package command

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TimeInc/ape-dev-rt/commons"
	"github.com/TimeInc/ape-dev-rt/git"
)

type Config struct {
	Variables map[string]interface{} `json:"variables"`
}

func CreateApp(c *commons.Context) error {
	skeleton := c.String("skeleton")
	appName := c.String("app")
	targetPath := c.String("path")

	repoPath, err := git.GetRepositoryPath()
	if err != nil {
		return err
	}

	err = copyTree(
		filepath.Join(repoPath, "skeletons", skeleton),
		filepath.Join(targetPath, appName))
	if err != nil {
		return err
	}

	variables, err := parseVariablesConfig(
		filepath.Join(repoPath, "skeletons", skeleton+".variables.json"))
	if err != nil {
		return err
	}

	vars, err := askForVariables(os.Stdin, os.Stdout, variables)
	if err != nil {
		return err
	}

	vars["AppName"] = appName
	_, err = commons.ProcessTemplatesAndDelete(
		filepath.Join(targetPath, appName), "tpl", vars)
	if err != nil {
		return err
	}

	return nil
}

func parseVariablesConfig(path string) (map[string]interface{}, error) {
	configContents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := Config{}
	log.Printf("[DEBUG] Unmarshalling %q", configContents)
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling %q: %q", path, err.Error())
	}

	return config.Variables, nil
}

func askForVariables(inputReader io.Reader, outputWriter io.Writer,
	variables map[string]interface{}) (map[string]string, error) {

	result := make(map[string]string, len(variables))
	reader := bufio.NewReader(inputReader)

	// To store the keys in slice in sorted order
	var keys []string
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		field := variables[k].(map[string]interface{})

		answer, err := askForVariable(reader, outputWriter, field)
		if err != nil {
			return nil, err
		}
		result[k] = answer
	}

	return result, nil
}

func askForVariable(reader *bufio.Reader, outputWriter io.Writer,
	field map[string]interface{}) (string, error) {

	if v, ok := field["default"]; ok && v != "" {
		fmt.Fprintf(outputWriter, "\n%s (%s): ", field["description"], v)
	} else {
		fmt.Fprintf(outputWriter, "\n%s: ", field["description"])
	}

	answer, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	answer = strings.TrimSpace(answer)

	if answer == "" {
		if v, ok := field["default"]; ok {
			return v.(string), nil
		}
		return askForVariable(reader, outputWriter, field)
	}

	return answer, nil
}

func copyTree(sourcePath, targetPath string) error {
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Source target path: %q", absSource)
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	return filepath.Walk(absSource, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Walk error: %q", err.Error())
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(absSource, path)
		if err != nil {
			return err
		}
		target := filepath.Join(absTarget, relPath)
		log.Printf("[DEBUG] Trying to copy a file from %q to %q", path, target)
		err = copyFile(path, target)
		if err != nil {
			return fmt.Errorf("Error copying file: %q", err.Error())
		}
		return nil
	})
}

func copyFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	sourceBaseDir := filepath.Dir(sourcePath)
	sourceInfo, err := os.Stat(sourceBaseDir)
	if err != nil {
		return err
	}

	targetBaseDir := filepath.Dir(targetPath)
	if _, err := os.Stat(targetBaseDir); os.IsNotExist(err) {
		log.Printf("[DEBUG] Creating base dir: %q", targetBaseDir)
		err := os.MkdirAll(targetBaseDir, sourceInfo.Mode())
		if err != nil {
			return err
		}
	}
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("Error creating file: %q", err.Error())
	}
	defer target.Close()

	bytesCopied, err := io.Copy(target, source)
	if err != nil {
		return fmt.Errorf("Error copying file: %q", err.Error())
	}
	log.Printf("Copied %s to %s (%d bytes)", sourcePath, targetPath, bytesCopied)
	return nil
}
