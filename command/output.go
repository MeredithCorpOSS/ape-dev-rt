package command

import (
	"fmt"
	"os"

	"github.com/MeredithCorpOSS/ape-dev-rt/aws"
	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
	"github.com/MeredithCorpOSS/ape-dev-rt/hcl"
	"github.com/MeredithCorpOSS/ape-dev-rt/terraform"
)

func Output(c *commons.Context) error {
	user, ok := c.CliContext.App.Metadata["user"].(*aws.User)
	if !ok {
		return fmt.Errorf("Unable to find AWS User in metadata")
	}
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}
	rs, ok := c.CliContext.App.Metadata["remote_state"].(*hcl.RemoteState)
	if !ok {
		return fmt.Errorf("Unable to find Remote State in metadata")
	}

	cfgPath, err := os.Getwd()
	if err != nil {
		return err
	}
	rootDir := cfgPath

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

	templateVars := ApplicationTemplateVars{
		AwsAccountId: namespace,
		AppName:      c.String("app"),
		Environment:  c.String("env"),
	}

	filesToCleanup := make([]string, 0)

	filesToCleanup, err = commons.ProcessTemplates(rootDir, "tpl", templateVars)
	if err != nil {
		return err
	}

	remoteState, err := terraform.GetRemoteStateForApp(&terraform.RemoteState{
		Backend: rs.Backend,
		Config:  rs.Config,
	}, namespace, c.String("app"))
	if err != nil {
		return err
	}

	filesToCleanup = append(filesToCleanup, terraform.GetBackendConfigFilename(rootDir))

	outputs, err := terraform.FreshOutput(remoteState, cfgPath)
	if err != nil {
		return err
	}

	outputMessage, err := generateOutputMessage(c.String("app"), c.String("env"), "", c.String("name"), outputs)
	if err != nil {
		return err
	}

	fmt.Println(outputMessage)

	return cleanupFilePaths(filesToCleanup)
}
