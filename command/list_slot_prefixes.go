package command

import (
	"fmt"

	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
)

func ListSlotPrefixes(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

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

	if len(appData.SlotCounters) == 0 {
		return fmt.Errorf("No slot counters found for %q in %q", c.String("app"), c.String("env"))
	}

	for prefix, counter := range appData.SlotCounters {
		fmt.Printf("%s\t%d\n", prefix, counter)
	}

	return nil
}
