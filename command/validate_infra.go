package command

import (
	"fmt"
	"os"

	"github.com/MeredithCorpOSS/ape-dev-rt/aws"
	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
	"github.com/MeredithCorpOSS/ape-dev-rt/terraform"
)

func ValidateInfra(c *commons.Context) error {
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
	filesToCleanup, err = commons.ProcessTemplates(cfgPath, "tpl", templateVars)
	if err != nil {
		return err
	}

	vo, err := terraform.Validate(cfgPath)
	if err != nil {
		return err
	}

	if vo.ExitCode != 0 {
		return fmt.Errorf("validate operation failed (exit code %d). Stderr:\n%s",
			vo.ExitCode, vo.Stderr)
	}

	fmt.Printf("\\(◕ヮ◕)/\nTerraform code for app '%s' in environment '%s' is valid.\n", c.String("app"), c.String("env"))

	return cleanupFilePaths(filesToCleanup)
}
