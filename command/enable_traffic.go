package command

import (
	"errors"
	"fmt"
	"log"

	"github.com/MeredithCorpOSS/ape-dev-rt/aws"
	"github.com/MeredithCorpOSS/ape-dev-rt/commons"
	"github.com/MeredithCorpOSS/ape-dev-rt/deploymentstate"
	"github.com/MeredithCorpOSS/ape-dev-rt/terraform"
)

func EnableTraffic(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	regionalAWS := aws.NewAWS(c.GlobalString("aws-profile"), c.String("aws-region"))
	fmt.Printf("Operating on resources in AWS region %s\n\n", colour.boldWhite(*regionalAWS.Region))
	user, err := regionalAWS.User()
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Received AWS Account: %#v", user)

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

	slotId := c.String("slot-id")
	slotPrefix := c.String("slot-prefix")
	if slotId == "" && slotPrefix == "" {
		return fmt.Errorf("'slot-id' or 'slot-prefix' is required parameter for %q (migrated app)", c.String("app"))
	}
	if slotId != "" && slotPrefix != "" {
		return errors.New("You can specify either 'slot-id' or 'slot-prefix', not both.")
	}

	if slotPrefix != "" {
		counter, prefixExists, err := ds.GetSlotCounter(slotPrefix, appData)
		if err != nil {
			return err
		}
		if !prefixExists {
			return fmt.Errorf("Slot prefix %s does not exist", slotPrefix)
		}

		slotId = fmt.Sprintf("%s%d", slotPrefix, counter)
		fmt.Printf("Enabling traffic for last slot (%s)\n", colour.boldWhite(slotId))
	}

	app, err := ds.GetApplication(c.String("app"))
	if err != nil {
		return err
	}
	if app.InfraOutputs == nil {
		return fmt.Errorf("No infra outputs found for %q", c.String("app"))
	}
	outputs := app.InfraOutputs

	internalAppName, ok := outputs[terraform.AppName]
	if !ok {
		return fmt.Errorf("Output %q not found", terraform.AppName)
	}

	var scalingGroup string
	scalingGroup, err = regionalAWS.GetScalingGroupForSlotId(c.String("env"), internalAppName, slotId)
	if err != nil {
		return fmt.Errorf("Failed getting scaling group for %s slot %s", internalAppName, slotId)
	}
	if len(scalingGroup) == 0 {
		return fmt.Errorf("Slot %s has no scaling group\n", slotId)
	}

	var balancers []string
	balancers, err = regionalAWS.GetBalancersForApp(internalAppName)
	if err != nil {
		return fmt.Errorf("Failed getting load balancers for %s", internalAppName)
	}

	if len(balancers) == 0 {
		return fmt.Errorf("No Load Balancer found for %s\n", internalAppName)
	}

	for _, b := range balancers {
		instances, err := regionalAWS.DescribeBalancedInstanceHealth(b)
		if err != nil {
			return fmt.Errorf("Failed asessing health of instances attached to %s", b)
		}
		if len(instances) > 0 {
			fmt.Printf("(ELB %s already has %d instances attached)\n", b, len(instances))
		}
	}

	err = regionalAWS.AttachBalancersToScalingGroup(balancers, scalingGroup)
	if err != nil {
		return fmt.Errorf("Failed attaching balancers %s, to scaling group %s", balancers, scalingGroup)
	}
	attachedNotice := fmt.Sprintf("Load Balancers attached to scaling group %s\n", scalingGroup)
	fmt.Printf("%s", colour.boldGreen(attachedNotice))
	return nil
}
