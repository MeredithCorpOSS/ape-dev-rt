package command

import (
	"fmt"
	"strings"

	"github.com/MeredithCorpOSS/ape-dev-rt/clippy"
	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate/schema"
)

func AddSlotPrefix(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	if c.CliContext.NArg() < 1 {
		return fmt.Errorf("You need to supply the name of slot prefix to create for %q / %q.",
			c.String("app"), c.String("env"))
	}

	prefix := strings.TrimSpace(c.CliContext.Args().First())

	appData, exists, err := BeginApplicationOperation(c.String("env"), c.String("app"), ds)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if appData.UseCentralGitRepo {
		return fmt.Errorf("You cannot use slot prefixes for %q as it comes from the central git repo",
			c.String("app"))
	}

	_, prefixExists, err := ds.GetSlotCounter(prefix, appData)
	if err != nil {
		return err
	}
	if prefixExists {
		return fmt.Errorf("Slot prefix %q already exists for %q / %q,",
			prefix, c.String("app"), c.String("env"))
	}

	note := fmt.Sprintf("It looks like you want to create a new prefix %q for %q in %q",
		prefix, c.String("app"), c.String("env"))
	isSensitive := isEnvironmentSensitive(c.String("env"))
	out, confirmed, err := clippy.BoolPrompt(note, c.Bool("y"), isSensitive, func() (interface{}, error) {
		return ds.AddSlotCounter(prefix, appData)
	}, nil)
	if err != nil {
		return err
	}
	if confirmed {
		appData = out.(*schema.ApplicationData)
		err = FinishApplicationOperation(c.String("app"), appData, appData.IsActive, nil, ds)
		if err != nil {
			return err
		}
		fmt.Printf(colour.boldGreen("Slot prefix %q created for %q / %q.\n"),
			prefix, c.String("app"), c.String("env"))
	}

	return nil
}
