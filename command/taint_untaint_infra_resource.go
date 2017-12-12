package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/TimeInc/ape-dev-rt/aws"
	"github.com/TimeInc/ape-dev-rt/commons"
	"github.com/TimeInc/ape-dev-rt/deploymentstate"
	"github.com/TimeInc/ape-dev-rt/hcl"
	"github.com/TimeInc/ape-dev-rt/terraform"
)

func TaintUntaintInfraResource(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}
	rs, ok := c.CliContext.App.Metadata["remote_state"].(*hcl.RemoteState)
	if !ok {
		return fmt.Errorf("Unable to find Remote State in metadata")
	}
	user, ok := c.CliContext.App.Metadata["user"].(*aws.User)
	if !ok {
		return fmt.Errorf("Unable to find AWS User in metadata")
	}

	namespace := user.AccountID
	if c.String("namespace") != "default" {
		namespace = c.String("namespace")
	}

	actionMap := map[string]string{
		"taint-infra-resource":   "taint",
		"untaint-infra-resource": "untaint",
	}

	action, ok := actionMap[c.CliContext.Command.Name]
	if !ok {
		return fmt.Errorf("Unexpected command name: %s", c.CliContext.Command.Name)
	}

	if c.CliContext.NArg() < 1 {
		return fmt.Errorf("You need to supply a resource to %s", action)
	}

	resource := c.CliContext.Args().First()

	cfgPath, err := os.Getwd()
	if err != nil {
		return err
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

	rootDir := cfgPath

	remoteState, err := terraform.GetRemoteStateForApp(&terraform.RemoteState{
		Backend: rs.Backend,
		Config:  rs.Config,
	}, user.AccountID, c.String("app"))
	if err != nil {
		return err
	}

	filesToCleanup := make([]string, 0)
	_, err = terraform.ReenableRemoteState(remoteState, rootDir)
	if err != nil {
		return err
	}

	args := []string{}
	var moduleMsg string
	if c.CliContext.IsSet("module") {
		module := fmt.Sprintf("-module=%s", c.String("module"))
		args = append(args, module)
		moduleMsg = fmt.Sprintf(" in module %s", c.String("module"))
	}
	args = append(args, "-no-color")
	args = append(args, resource)

	fmt.Printf("%sing infra resource %s%s for application %s in %s/%s.\n",
		strings.Title(action), resource, moduleMsg, c.String("app"), namespace, c.String("env"))

	out, err := terraform.Cmd(action, args, rootDir, ioutil.Discard, ioutil.Discard)
	if err != nil {
		return err
	}
	if out.ExitCode != 0 {
		return fmt.Errorf("Failed to %s a resource (exit code %d). Stderr:\n%s",
			action, out.ExitCode, out.Stderr)
	}
	filesToCleanup = append(filesToCleanup, terraform.GetBackendConfigFilename(rootDir))

	fmt.Printf("%s\n", out.Stdout)

	return cleanupFilePaths(filesToCleanup)

}
