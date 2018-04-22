package command

import (
	"fmt"

	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
)

var whitelistedOutputs = map[string]bool{
	"app":     true,
	"lb_fqdn": true,
}

func ListApps(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	fmt.Printf("%s Listing apps from the main deploymentstate backend.\n", colour.boldWhite("Note:"))

	apps, err := ds.ListApplications()
	if err != nil {
		return err
	}
	if len(apps) == 0 {
		fmt.Println(colour.boldWhite("No applications found."), " Did you eat them all?! :o\n")
		return nil
	}
	for _, a := range apps {
		if a.IsActive {
			fmt.Printf("%s", colour.boldWhite(a.Name))

			legacySuffix := ""
			if a.UseCentralGitRepo {
				legacySuffix = ", " + colour.boldYellow("uses central git repo")
			}
			fmt.Printf(" (RT %s%s)\n", a.LastRtVersion, legacySuffix)
			if a.InfraOutputs != nil && len(a.InfraOutputs) > 0 {
				for k, v := range a.InfraOutputs {
					if whitelistedOutputs[k] {
						fmt.Printf(" - %s: %s\n", k, v)
					}
				}
			}
		} else {
			fmt.Printf("%s (%s)", a.Name, colour.red("inactive"))
		}
		fmt.Println("")
	}
	fmt.Println("")

	return nil
}
