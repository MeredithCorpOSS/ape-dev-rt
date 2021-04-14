package command

import (
	"fmt"

	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
)

func ListSlots(c *commons.Context) error {
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

	for _, s := range slots {
		prefix := ""
		finishTime := ""
		if s.LastTerraformRun == nil {
			fmt.Printf("%s\n", colour.boldWhite(s.SlotId))
			prefix = colour.boldYellow("being deployed")
		} else if s.IsActive {
			fmt.Printf("%s\n", colour.boldWhite(s.SlotId))
			prefix = "last " + colour.green("deployed")
			finishTime = s.LastTerraformRun.FinishTime.String()
		} else {
			fmt.Printf("%s\n", s.SlotId)
			prefix = colour.red("destroyed")
			finishTime = s.LastTerraformRun.FinishTime.String()
		}

		pilotSuffix := ""
		if s.LastDeployPilot != nil {
			pilotSuffix = fmt.Sprintf("by %s via %s",
				s.LastDeployPilot.AWSApiCaller,
				s.LastDeployPilot.IPAddress)
		}
		fmt.Printf(" - %s %s %s\n",
			prefix,
			finishTime,
			pilotSuffix)

		if s.LastTerraformRun != nil && len(s.LastTerraformRun.Variables) > 0 {
			fmt.Printf(" - last variables: %q\n", s.LastTerraformRun.Variables)
		}
		if s.LastTerraformRun != nil && len(s.LastTerraformRun.Outputs) > 0 {
			fmt.Printf(" - last outputs: %q\n", s.LastTerraformRun.Outputs)
		}
		fmt.Println("")
	}

	return nil
}
