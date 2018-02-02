package command

import (
	"fmt"
	"os"
	"path"

	"github.com/TimeInc/ape-dev-rt/aws"
	"github.com/TimeInc/ape-dev-rt/commons"
	"github.com/TimeInc/ape-dev-rt/deploymentstate"
	"github.com/TimeInc/ape-dev-rt/hcl"
	"github.com/TimeInc/ape-dev-rt/terraform"
)

func SlotOutput(c *commons.Context) error {
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

	slotId := c.String("slot-id")
	if slotId == "" {
		return fmt.Errorf("Please provide a slot-id for %q in environment %q", c.String("app"), c.String("env"))
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

	if c.CliContext.NArg() < 1 {
		return fmt.Errorf("You need to supply a path to Terraform configs of %q.", slotId)
	}

	_, err = ds.GetSlot(c.String("app"), slotId)
	if err != nil {
		return err
	}

	namespace := user.AccountID
	if c.String("namespace") != "default" {
		namespace = c.String("namespace")
	}

	rootDir := path.Join(cfgPath, c.CliContext.Args().First())

	remoteState, err := terraform.GetRemoteStateForSlotId(&terraform.RemoteState{
		Backend: rs.Backend,
		Config:  rs.Config,
	}, namespace, c.String("app"), slotId)
	if err != nil {
		return err
	}

	outputs, err := terraform.FreshOutput(remoteState, rootDir)
	if err != nil {
		return err
	}

	outputMessage, err := generateOutputMessage(c.String("app"), c.String("env"), slotId, c.String("name"), outputs)
	if err != nil {
		return err
	}

	fmt.Println(outputMessage)

	return nil
}
