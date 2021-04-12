package command

import (
	"fmt"

	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
)

func ListDeployments(c *commons.Context) error {
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
		return deprecatedGitError()
	}

	slots, err := ds.ListSlots(c.String("app"))
	if err != nil {
		return err
	}

	for _, s := range slots {
		if s.IsActive {
			pilotSuffix := ""
			if s.LastDeployPilot != nil {
				pilotSuffix = fmt.Sprintf(" (last %s by %s)",
					colour.green("deployed"),
					s.LastDeployPilot.AWSApiCaller)
			}
			fmt.Printf("%s%s",
				colour.boldWhite(s.SlotId),
				pilotSuffix)
			deployments, err := ds.ListLastDeployments(c.String("app"), s.SlotId, 10)
			if err != nil {
				return err
			}

			if len(deployments) > 0 {
				fmt.Println("")
				for _, d := range deployments {
					tfAction := colour.boldGreen("apply")
					if d.Terraform.IsDestroy {
						tfAction = colour.boldRed("destroy")
					}

					suffix := ""
					if !d.StartTime.IsZero() {
						suffix += fmt.Sprintf(" started %s", d.StartTime)
					}
					if d.DeployPilot != nil {
						suffix += fmt.Sprintf(" by %s via %s",
							d.DeployPilot.AWSApiCaller,
							d.DeployPilot.IPAddress)
					}

					fmt.Printf(" - %s (%s%s)\n",
						d.DeploymentId,
						tfAction,
						suffix)

					fmt.Printf("   - finished: %s\n", d.Terraform.FinishTime)
					fmt.Printf("   - variables: %q\n", d.Terraform.Variables)
					fmt.Printf("   - outputs: %q\n", d.Terraform.Outputs)
					exitCode := fmt.Sprintf("%d", d.Terraform.ExitCode)
					if exitCode != "0" {
						exitCode = colour.boldRed(fmt.Sprintf("%s (!)", exitCode))
						fmt.Printf("   - exit code: %s\n", exitCode)
					}
				}
			}
		} else {
			pilotSuffix := ""
			if s.LastDeployPilot != nil {
				pilotSuffix = fmt.Sprintf(" by %s",
					s.LastDeployPilot.AWSApiCaller)
			}
			fmt.Printf("%s (%s %s%s)",
				s.SlotId,
				colour.red("destroyed"),
				s.LastTerraformRun.FinishTime,
				pilotSuffix)
		}
		fmt.Println("")
	}
	fmt.Println("")

	return nil
}
