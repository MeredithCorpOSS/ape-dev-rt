package command

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/TimeIncOSS/ape-dev-rt/clippy"
	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/schema"
)

func DeleteSlotPrefix(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	if c.CliContext.NArg() < 1 {
		return fmt.Errorf("You need to supply the name of slot prefix to delete for %q / %q.",
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

	if !prefixExists {
		return fmt.Errorf("Slot prefix %q does not exist for %q / %q,",
			prefix, c.String("app"), c.String("env"))
	}

	slotIdMatches, err := countMatchingSlotIds(ds, prefix, c.String("app"))
	if err != nil {
		return err
	}

	note := generateNote(slotIdMatches, prefix, c.String("app"), c.String("env"))
	isSensitive := isEnvironmentSensitive(c.String("env"))
	out, confirmed, err := clippy.BoolPrompt(note, c.Bool("y"), isSensitive, func() (interface{}, error) {
		return ds.DeleteSlotCounter(prefix, appData)
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
		fmt.Printf(colour.boldGreen("Slot prefix %q deleted for %q / %q.\n"),
			prefix, c.String("app"), c.String("env"))
	}

	return nil
}

func countMatchingSlotIds(ds *deploymentstate.DeploymentState, prefix, appName string) (int, error) {
	slotIdMatches := 0
	slotData, err := ds.ListSlots(appName)
	if err != nil {
		return slotIdMatches, err
	}
	for _, slot := range slotData {
		match, err := regexp.MatchString(prefix, slot.SlotId)
		if err != nil {
			return 0, err
		}
		if match && slot.IsActive {
			slotIdMatches++
		}
	}
	return slotIdMatches, nil
}

func generateNote(slotIdMatches int, prefix, appName, env string) string {
	if slotIdMatches == 1 {
		return fmt.Sprintf("The prefix %q matches %d active Slot ID for %q in %q. Are you sure you want to continue?",
			prefix, slotIdMatches, appName, env)
	}
	if slotIdMatches > 1 {
		return fmt.Sprintf("The prefix %q matches %d active Slot IDs for %q in %q. Are you sure you want to continue?",
			prefix, slotIdMatches, appName, env)
	}
	return fmt.Sprintf("It looks like you want to delete the prefix %q for %q in %q",
		prefix, appName, env)
}
