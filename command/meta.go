package command

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/MeredithCorpOSS/ape-dev-rt/clippy"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate/backends"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate/schema"
	"github.com/MeredithCorpOSS/ape-dev-rt/rt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/ttacon/chalk"
)

var dateLayout = "Mon Jan 2 15:04:05 -0700 2006"

type chalkFunc func(string) string

type colours struct {
	boldGreen  chalkFunc
	boldWhite  chalkFunc
	boldYellow chalkFunc
	boldRed    chalkFunc
	boldBlue   chalkFunc
	red        chalkFunc
	green      chalkFunc
}

var colour = &colours{
	boldGreen:  chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style,
	boldWhite:  chalk.White.NewStyle().WithTextStyle(chalk.Bold).Style,
	boldYellow: chalk.Yellow.NewStyle().WithTextStyle(chalk.Bold).Style,
	boldRed:    chalk.Red.NewStyle().WithTextStyle(chalk.Bold).Style,
	boldBlue:   chalk.Blue.NewStyle().WithTextStyle(chalk.Bold).Style,
	red:        chalk.Red.NewStyle().Style,
	green:      chalk.Green.NewStyle().Style,
}

type ApplicationTemplateVars struct {
	AwsAccountId, Environment, AppName string
}

type VersionTemplateVars struct {
	AwsAccountId, Environment, AppName string
}

func turnOffColours(c *colours) {
	c.boldGreen = chalk.Reset.Style
	c.boldWhite = chalk.Reset.Style
	c.boldYellow = chalk.Reset.Style
	c.boldRed = chalk.Reset.Style
	c.boldBlue = chalk.Reset.Style
	c.red = chalk.Reset.Style
	c.green = chalk.Reset.Style
}

func BeginApplicationOperation(env, appName string, ds *deploymentstate.DeploymentState, yesOverride ...bool) (*schema.ApplicationData, bool, error) {
	yes := false // use this varaible to control output of creation prompt
	if len(yesOverride) > 0 {
		yes = yesOverride[0]
	}
	app, err := ds.GetApplication(appName)
	if err != nil {
		_, ok := err.(*backends.AppNotFound)
		if ok {
			note := fmt.Sprintf("Application %q doesn't exist in %q, do you want to create it?", appName, env)
			isSensitive := isEnvironmentSensitive(env)
			out, confirmed, _ := clippy.BoolPrompt(note, yes, isSensitive, func() (interface{}, error) {
				return &schema.ApplicationData{
					UseCentralGitRepo:    false,
					LastRtVersion:        rt.Version,
					LastTerraformVersion: rt.TerraformVersion,
					IsActive:             true,
				}, nil
			}, nil)
			if !confirmed {
				return nil, false, nil
			}
			app = out.(*schema.ApplicationData)
		} else {
			return nil, false, err
		}
	}

	lastRtVersion, err := version.NewVersion(app.LastRtVersion)
	if err != nil {
		return nil, false, err
	}
	currentRtVersion, err := version.NewVersion(rt.Version)
	if err != nil {
		return nil, false, err
	}
	if currentRtVersion.LessThan(lastRtVersion) {
		return nil, false, fmt.Errorf("Last used RT version for %q: %s. You have %q, please upgrade.",
			appName, app.LastRtVersion, rt.Version)
	}

	return app, true, nil
}

func FinishApplicationOperation(appName string, appData *schema.ApplicationData, isActive bool, outputs map[string]string,
	ds *deploymentstate.DeploymentState) error {
	appData.IsActive = isActive
	appData.LastRtVersion = rt.Version
	appData.LastTerraformVersion = rt.TerraformVersion
	if outputs != nil && len(outputs) > 0 {
		appData.InfraOutputs = outputs
	}

	err := ds.SaveApplication(appName, appData)
	return err
}

func isEnvironmentSensitive(environment string) bool {
	re := regexp.MustCompile("(prod|production|live)")
	return re.Match([]byte(environment))
}

func cleanupFilePaths(paths []string) error {
	var errors error
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			errors = multierror.Append(err)
		}
	}
	return errors
}

func generateOutputMessage(app, env, slotId, name string, outputs map[string]string) (string, error) {
	slotIdMessage := ""
	if slotId != "" {
		slotIdMessage = fmt.Sprintf(" with slot-id %s", slotId)
	}
	if name == "" {
		b, err := json.MarshalIndent(outputs, "", "	")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("The app %s%s in env %s contains the outputs:\n%s\n", app, slotIdMessage, env, b), nil
	}

	value, ok := outputs[name]

	if ok {
		return fmt.Sprintf("The app %s%s in env %s contains the output:\n%s: %s\n", app, slotIdMessage, env, name, value), nil
	}

	return "", fmt.Errorf("The app %s%s in env %s does not contain the output:\n%s\n", app, slotIdMessage, env, name)
}

func deprecatedGitError() error {
	return fmt.Errorf("%s", `
                           ____________________________________
     ___      ___         /                                    \
    /   \____/   \        | Please migrate your app out of the |
   /    / \/ \    \       | central Git repository as it is    |
  /    |  ..  |    \   <--| deprecated.                        |
  \___/|      |\___/\     \____________________________________/
     | |_|  |_|      \
     | |/|__|\|       \
     |   |__|         |\
     |   |__|   |_/  /  \
     | @ |  | @ || @ |   '
     |   |~~|   ||   |
     'ooo'  'ooo''ooo'
   `)
}
