package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/TimeInc/ape-dev-rt/aws"
	"github.com/TimeInc/ape-dev-rt/clippy"
	"github.com/TimeInc/ape-dev-rt/commons"
	"github.com/TimeInc/ape-dev-rt/deploymentstate"
	"github.com/TimeInc/ape-dev-rt/deploymentstate/schema"
	"github.com/TimeInc/ape-dev-rt/hcl"
	"github.com/TimeInc/ape-dev-rt/terraform"
)

func DeployDestroy(c *commons.Context) error {
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
	currentIp, ok := c.CliContext.App.Metadata["current_ip"].(string)
	if !ok {
		fmt.Print(colour.boldYellow("Note: We were unable to detect your IP address\n"))
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
	slotPrefix := c.String("slot-prefix")
	if slotId == "" && slotPrefix == "" {
		return fmt.Errorf("'slot-id' or 'slot-prefix' is required parameter for %q (migrated app)", c.String("app"))
	}
	if slotId != "" && slotPrefix != "" {
		return errors.New("You can specify either 'slot-id' or 'slot-prefix', not both.")
	}

	if slotPrefix != "" {
		counter, prefixExists, err := ds.GetSlotCounter(slotPrefix, appData)
		if err != nil {
			return err
		}
		if !prefixExists {
			return fmt.Errorf("Slot prefix %s does not exist", slotPrefix)
		}

		if c.Bool("previous-slot") {
			lastSlotId := fmt.Sprintf("%s%d", slotPrefix, counter)
			counter -= 1
			slotId = fmt.Sprintf("%s%d", slotPrefix, counter)
			fmt.Printf("Last slot ID is %s, preparing destruction of %s\n", lastSlotId, colour.boldRed(slotId))
		} else {
			slotId = fmt.Sprintf("%s%d", slotPrefix, counter)
			fmt.Printf("Preparing destruction of slot ID %s\n", colour.boldRed(slotId))
		}
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

	filesToCleanup := make([]string, 0)
	templateVars := VersionTemplateVars{
		AwsAccountId: namespace,
		Environment:  c.String("env"),
		AppName:      c.String("app"),
	}
	filesToCleanup, err = commons.ProcessTemplates(rootDir, "tpl", templateVars)
	if err != nil {
		return err
	}

	slotData, err := ds.GetSlot(c.String("app"), slotId)
	if err != nil {
		return err
	}
	tfVariables := slotData.LastTerraformRun.Variables
	vars := c.StringSlice("var")
	for _, v := range vars {
		parts := strings.Split(v, "=")
		key, value := parts[0], parts[1]
		tfVariables[key] = value
	}
	tfVariables["app_name"] = c.String("app")
	tfVariables["app_version"] = slotId
	tfVariables["environment"] = c.String("env")

	remoteState, err := terraform.GetRemoteStateForSlotId(&terraform.RemoteState{
		Backend: rs.Backend,
		Config:  rs.Config,
	}, user.AccountID, c.String("app"), slotId)
	if err != nil {
		return err
	}

	planStartTime := time.Now().UTC()
	filesToCleanup = append(filesToCleanup, path.Join(rootDir, ".terraform"))
	out, err := terraform.FreshPlan(&terraform.FreshPlanInput{
		RemoteState: remoteState,
		RootPath:    rootDir,
		Variables:   tfVariables,
		Refresh:     true,
		Target:      c.String("target"),
		Destroy:     true,
	})
	planFinishTime := time.Now().UTC()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Plan started %s, finished %s",
		planStartTime.String(), planFinishTime.String())

	if out.ExitCode != 0 {
		return fmt.Errorf("Planning failed (exit code %d). Stderr:\n%s",
			out.ExitCode, out.Stderr)
	}
	diff := out.Diff
	if diff.ToChange+diff.ToCreate+diff.ToRemove == 0 && !c.Bool("f") {
		return cleanupFilePaths(filesToCleanup)
	}

	note := fmt.Sprintf(
		"It looks like you want to DESTROY SLOT %s of '%s' (%s/%s).",
		slotId, c.String("app"), namespace,
		c.String("env"))
	yesOverride := c.Bool("y")
	isSensitive := isEnvironmentSensitive(c.String("env"))
	var destroyStartTime time.Time
	var data *schema.DeploymentData
	destroyOut, confirmed, err := clippy.BoolPrompt(note, yesOverride, isSensitive, func() (interface{}, error) {
		destroyStartTime = time.Now().UTC()
		pilot := &schema.DeployPilot{
			AWSApiCaller: user.Arn,
			IPAddress:    currentIp,
		}
		var err error
		startTime := time.Now().UTC()
		data, err = ds.BeginDeployment(c.String("app"), slotId, true, pilot, startTime, tfVariables)
		if err != nil {
			return nil, err
		}

		input := terraform.DestroyInput{
			RootPath:     rootDir,
			Target:       "",
			Refresh:      true,
			Variables:    tfVariables,
			StderrWriter: ioutil.Discard,
		}
		return terraform.Destroy(&input)
	}, nil)
	if !confirmed {
		log.Printf("[DEBUG] User didn't confirm clippy dialog - not destroying.")
		return cleanupFilePaths(filesToCleanup)
	}
	if err != nil {
		return fmt.Errorf("Destroy operation failed: %s", err)
	}

	do := destroyOut.(*terraform.DestroyOutput)

	isActive := true
	isStateEmpty, err := terraform.IsStateEmpty(rootDir)
	if err != nil {
		return err
	}
	isActive = !isStateEmpty

	err = ds.FinishDeployment(c.String("app"), slotId, data.DeploymentId, isActive, data, &schema.FinishedTerraformRun{
		PlanStartTime:  planStartTime,
		PlanFinishTime: planFinishTime,
		StartTime:      destroyStartTime,
		FinishTime:     time.Now().UTC(),
		ResourceDiff:   do.Diff,
		ExitCode:       do.ExitCode,
		Warnings:       do.Warnings,
		Stderr:         do.Stderr,
	})
	if err != nil {
		return fmt.Errorf("Finishing deployment failed: %s", err)
	}

	appData.LastDeploymentTime = time.Now().UTC()
	err = FinishApplicationOperation(c.String("app"), appData, true, nil, ds)
	if err != nil {
		return err
	}

	if isStateEmpty {
		if do.ExitCode != 0 {
			fmt.Printf("%s\n\n", colour.boldWhite("Slot was successfully destroyed. Terraform encountered some errors."))
		}
	} else {
		if do.ExitCode != 0 {
			fmt.Printf("%s\n\n", colour.boldWhite("Slot could not be destroyed because Terraform encountered some errors."))
		}
	}

	if do.ExitCode != 0 {
		return fmt.Errorf("Errors:\n%s", do.Stderr)
	}

	return cleanupFilePaths(filesToCleanup)
}
