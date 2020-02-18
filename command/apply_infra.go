package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/TimeIncOSS/ape-dev-rt/aws"
	"github.com/TimeIncOSS/ape-dev-rt/clippy"
	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/hcl"
	"github.com/TimeIncOSS/ape-dev-rt/terraform"
)

func ApplyInfra(c *commons.Context) error {
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

	namespace := user.AccountID
	if c.String("namespace") != "default" {
		namespace = c.String("namespace")
	}

	rootDir := cfgPath
	filesToCleanup := make([]string, 0)
	templateVars := ApplicationTemplateVars{
		AwsAccountId: namespace,
		AppName:      c.String("app"),
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
	planFilePath := path.Join(rootDir, "planfile")
	filesToCleanup = append(filesToCleanup, path.Join(rootDir, ".terraform"))
	filesToCleanup = append(filesToCleanup, path.Join(rootDir, "terraform.tfstate.backup"))
	filesToCleanup = append(filesToCleanup, planFilePath)
	planOut, err := terraform.FreshPlan(&terraform.FreshPlanInput{
		RemoteState:  remoteState,
		RootPath:     rootDir,
		PlanFilePath: planFilePath,
		Variables:    tfVariables,
		Refresh:      true,
		Target:       c.String("target"),
		Destroy:      false,
		XLegacy:      c.Bool("x"),
	})
	filesToCleanup = append(filesToCleanup, terraform.GetBackendConfigFilename(rootDir))
	planFinishTime := time.Now().UTC()
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Plan started %s, finished %s",
		planStartTime.String(), planFinishTime.String())

	if planOut.ExitCode != 0 {
		return fmt.Errorf("Plan failed (exit code %d). Stderr:\n%v", planOut.ExitCode, planOut.Stderr)
	}

	diff := planOut.Diff
	if diff.ToChange+diff.ToCreate+diff.ToRemove == 0 && !c.Bool("f") {
		return cleanupFilePaths(filesToCleanup)
	}

	note := fmt.Sprintf(
		"It looks like you want to change infrastructure of '%s' in %s/%s.",
		c.String("app"), namespace, c.String("env"))
	yesOverride := c.Bool("y")
	isSensitive := isEnvironmentSensitive(c.String("env"))
	applyOut, confirmed, err := clippy.BoolPrompt(note, yesOverride, isSensitive, func() (interface{}, error) {
		input := terraform.ApplyInput{
			RootPath:     rootDir,
			Target:       "",
			XLegacy:      c.Bool("x"),
			Refresh:      true,
			PlanFilePath: planFilePath,
			StderrWriter: ioutil.Discard,
		}
		return terraform.Apply(&input)
	}, nil)
	if err != nil {
		return err
	}
	if !confirmed {
		return cleanupFilePaths(filesToCleanup)
	}

	ao := applyOut.(*terraform.ApplyOutput)
	log.Printf("[DEBUG] Apply done: %#v", ao)

	isActive := true
	isStateEmpty, err := terraform.IsStateEmpty(rootDir)
	if err != nil {
		return err
	}
	isActive = !isStateEmpty

	log.Printf("[DEBUG] Marking app as active? %t", isActive)

	appData.LastInfraChangeTime = time.Now().UTC()
	err = FinishApplicationOperation(c.String("app"), appData, isActive, ao.Outputs, ds)
	if err != nil {
		return err
	}

	if ao.ExitCode != 0 {
		return fmt.Errorf("Apply operation failed (exit code %d). Stderr:\n%s",
			ao.ExitCode, ao.Stderr)
	}

	return cleanupFilePaths(filesToCleanup)
}
