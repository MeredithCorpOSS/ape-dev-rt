package command

import (
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
	"github.com/TimeInc/ape-dev-rt/hcl"
	"github.com/TimeInc/ape-dev-rt/terraform"
)

func DestroyInfra(c *commons.Context) error {
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

	slotData, err := ds.ListSlots(c.String("app"))
	if err != nil {
		return err
	}

	activeSlots := make([]string, 0)
	for _, slot := range slotData {
		if slot.IsActive {
			activeSlots = append(activeSlots, slot.SlotId)
		}
	}

	if len(activeSlots) > 0 {
		return fmt.Errorf(
			"Bailing out of destroy-infra...\n"+
				"Cannot destroy app %s in env %s while slots are active. "+
				"Please use deploy-destroy on active slots:\n%v",
			c.String("app"), c.String("env"), activeSlots)
	}

	namespace := user.AccountID
	if c.String("namespace") != "default" {
		namespace = c.String("namespace")
	}

	rootDir := cfgPath
	filesToCleanup := make([]string, 0)
	templateVars := ApplicationTemplateVars{
		AwsAccountId: namespace,
		Environment:  c.String("env"),
	}
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

	vars := c.StringSlice("var")
	tfVariables := make(map[string]string, 0)
	for _, v := range vars {
		parts := strings.Split(v, "=")
		key, value := parts[0], parts[1]
		tfVariables[key] = value
	}

	tfVariables["app_name"] = c.String("app")
	tfVariables["environment"] = c.String("env")

	planStartTime := time.Now().UTC()
	filesToCleanup = append(filesToCleanup, path.Join(rootDir, ".terraform"))
	filesToCleanup = append(filesToCleanup, path.Join(rootDir, "terraform.tfstate.backup"))
	planOut, err := terraform.FreshPlan(&terraform.FreshPlanInput{
		RemoteState: remoteState,
		RootPath:    rootDir,
		Variables:   tfVariables,
		Refresh:     true,
		Target:      c.String("target"),
		Destroy:     true,
		XLegacy:     c.Bool("x"),
	})
	filesToCleanup = append(filesToCleanup, terraform.GetBackendConfigFilename(rootDir))
	planFinishTime := time.Now().UTC()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Plan started %s, finished %s",
		planStartTime.String(), planFinishTime.String())

	if planOut.ExitCode != 0 {
		return fmt.Errorf("Planning failed (exit code %d). Stderr:\n", planOut.ExitCode, planOut.Stderr)
	}

	diff := planOut.Diff
	if diff.ToChange+diff.ToCreate+diff.ToRemove == 0 && !c.Bool("f") {
		return cleanupFilePaths(filesToCleanup)
	}

	note := fmt.Sprintf(
		"It looks like you want to DESTROY INFRASTRUCTURE of '%s' (%s/%s).",
		c.String("app"), namespace, c.String("env"))
	yesOverride := c.Bool("y")
	isSensitive := isEnvironmentSensitive(c.String("env"))
	var destroyStartTime time.Time
	destroyOut, confirmed, err := clippy.BoolPrompt(note, yesOverride, isSensitive, func() (interface{}, error) {
		destroyStartTime = time.Now().UTC()
		input := terraform.DestroyInput{
			RootPath:     rootDir,
			Target:       "",
			XLegacy:      c.Bool("x"),
			Variables:    tfVariables,
			Refresh:      true,
			StderrWriter: ioutil.Discard,
		}
		return terraform.Destroy(&input)
	}, nil)
	if !confirmed {
		log.Printf("[DEBUG] User didn't confirm clippy dialog - not destroying infra.")
		return cleanupFilePaths(filesToCleanup)
	}
	if err != nil {
		return fmt.Errorf("Destroy operation failed: %s", err)
	}

	isActive := true
	isStateEmpty, err := terraform.IsStateEmpty(rootDir)
	if err != nil {
		return err
	}
	isActive = !isStateEmpty

	appData.LastInfraChangeTime = time.Now().UTC()
	err = FinishApplicationOperation(c.String("app"), appData, isActive, nil, ds)
	if err != nil {
		return err
	}

	do := destroyOut.(*terraform.DestroyOutput)
	if do.ExitCode != 0 {
		return fmt.Errorf("Destroy operation failed (exit code %d). Stderr:\n%s",
			do.ExitCode, do.Stderr)
	}

	return nil
}
