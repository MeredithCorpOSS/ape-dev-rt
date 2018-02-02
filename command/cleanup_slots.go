package command

import (
	"fmt"
	"time"

	"github.com/TimeInc/ape-dev-rt/commons"
	"github.com/TimeInc/ape-dev-rt/deploymentstate"
	"github.com/ninibe/bigduration"
)

func CleanupSlots(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	_, exists, err := BeginApplicationOperation(c.String("env"), c.String("app"), ds)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	slots, err := ds.ListSlots(c.String("app"))
	if err != nil {
		return err
	}

	isVerbose := c.Bool("verbose")
	d, err := bigduration.ParseBigDuration(c.String("older-than"))
	if err != nil {
		return err
	}
	borderline := time.Now().Add(-1 * d.Duration())

	for _, s := range slots {
		if s.LastTerraformRun == nil {
			if isVerbose {
				fmt.Printf("Skipping %q as it is %s.\n", s.SlotId, colour.boldYellow("being deployed"))
			}
			continue
		}
		if s.IsActive {
			if isVerbose {
				fmt.Printf("Skipping %q as it is %s.\n", s.SlotId, colour.green("active"))
			}
			continue
		}
		if !s.LastDeploymentStartTime.Before(borderline) {
			if isVerbose {
				realDuration, _ := bigduration.ParseBigDuration(
					time.Since(s.LastDeploymentStartTime).String())
				fmt.Printf("Skipping %q as it is newer than %s (%s)\n",
					s.SlotId, c.String("older-than"), realDuration.Compact())
			}
			continue
		}

		fmt.Printf("Deleting %s ...", colour.boldWhite(s.SlotId))
		err := ds.DeleteSlot(c.String("app"), s.SlotId)
		if err != nil {
			fmt.Println("")
			return err
		}
		fmt.Println("DONE")
	}

	return nil
}
