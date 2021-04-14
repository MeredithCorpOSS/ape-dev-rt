package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/MeredithCorpOSS/ape-dev-rt/aws"
	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
	"github.com/MeredithCorpOSS/ape-dev-rt/hcl"
	"github.com/MeredithCorpOSS/ape-dev-rt/terraform"
)

func TaintUntaintDeployedResource(c *commons.Context) error {
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

	slotId := c.String("slot-id")
	if slotId == "" {
		return fmt.Errorf("'slot-id' is required parameter for %q (migrated app)", c.String("app"))
	}

	if c.CliContext.NArg() < 1 {
		return fmt.Errorf("You need to supply a path to Terraform configs of %q.", slotId)
	}

	rootDir := path.Join(cfgPath, c.CliContext.Args().First())
	if rootDir == cfgPath {
		return fmt.Errorf("Terraform configs for a slot have to be in a separate dir, not in %q!", cfgPath)
	}
	_, err = os.Stat(rootDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("%q does not exist", rootDir)
	}

	namespace := user.AccountID
	if c.String("namespace") != "default" {
		namespace = c.String("namespace")
	}

	actionMap := map[string]string{
		"taint-deployed-resource":   "taint",
		"untaint-deployed-resource": "untaint",
	}

	action, ok := actionMap[c.CliContext.Command.Name]
	if !ok {
		return fmt.Errorf("Unexpected command name: %s", c.CliContext.Command.Name)
	}

	if c.CliContext.NArg() < 2 {
		return fmt.Errorf("You need to supply a resource to %s", action)
	}

	resource := c.CliContext.Args().Get(1)

	remoteState, err := terraform.GetRemoteStateForSlotId(&terraform.RemoteState{
		Backend: rs.Backend,
		Config:  rs.Config,
	}, user.AccountID, c.String("app"), slotId)
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
		module := fmt.Sprintf("module.%s.%s", c.String("module"), resource)
		args = append(args, module)
		moduleMsg = fmt.Sprintf(" in module %s", c.String("module"))
	}

	fmt.Printf("%sing resource %s%s for application %s (slot %s) in %s/%s.\n",
		strings.Title(action), resource, moduleMsg, c.String("app"), slotId, namespace, c.String("env"))

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
