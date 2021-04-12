package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/TimeIncOSS/ape-dev-rt/aws"
	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/terraform"
)

func ValidateSlots(c *commons.Context) error {
	user, ok := c.CliContext.App.Metadata["user"].(*aws.User)
	if !ok {
		return fmt.Errorf("Unable to find AWS User in metadata")
	}

	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	cfgPath, err := os.Getwd()
	if err != nil {
		return err
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("%q does not exist", cfgPath)
	}

	files, err := ioutil.ReadDir(cfgPath)
	if err != nil {
		return err
	}

	if len(files) < 1 {
		return fmt.Errorf("Current working directory %s is empty", cfgPath)
	}

	directories := make([]string, 0)
	for _, file := range files {
		if file.Mode().IsDir() && !strings.HasPrefix(file.Name(), ".") {
			directories = append(directories, file.Name())
		}
	}

	appData, exists, err := BeginApplicationOperation(c.String("env"), c.String("app"), ds)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if appData.UseCentralGitRepo {
		return deprecatedGitError()
	}

	namespace := user.AccountID
	if c.String("namespace") != "default" {
		namespace = c.String("namespace")
	}

	filesToCleanup := make([]string, 0)
	templateVars := VersionTemplateVars{
		AwsAccountId: namespace,
		Environment:  c.String("env"),
		AppName:      c.String("app"),
	}

	for _, directory := range directories {
		rootDir := path.Join(cfgPath, directory)
		filesToCleanup, err = commons.ProcessTemplates(rootDir, "tpl", templateVars)
		if err != nil {
			return err
		}

		vo, err := terraform.Validate(rootDir)
		if err != nil {
			return err
		}
		if vo.ExitCode != 0 {
			return fmt.Errorf("validate operation failed (exit code %d). Stderr:\n%s",
				vo.ExitCode, vo.Stderr)
		}
	}

	fmt.Printf("\\(◕ヮ◕)/\nTerraform code for app '%s' in environment '%s' is valid.\n\n", c.String("app"), c.String("env"))

	return cleanupFilePaths(filesToCleanup)
}
