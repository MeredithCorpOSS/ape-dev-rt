package command

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/MeredithCorpOSS/ape-dev-rt/aws"
	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/hcl"
	"github.com/MeredithCorpOSS/ape-dev-rt/terraform"
)

func DiffInfra(c *commons.Context) error {
	user, ok := c.CliContext.App.Metadata["user"].(*aws.User)
	if !ok {
		return fmt.Errorf("Unable to find AWS User in metadata")
	}

	rs, ok := c.CliContext.App.Metadata["remote_state"].(*hcl.RemoteState)
	if !ok {
		return fmt.Errorf("Unable to find Remote State in metadata")
	}

	cfgPath, err := os.Getwd()
	if err != nil {
		return err
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

	return cleanupFilePaths(filesToCleanup)
}
